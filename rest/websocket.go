package rest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.qredo.com/custody-engine/automated-approver/defs"
)

const (
	// writeWait is the time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// pongWait is the time allowed to read the next pong message from the peer
	pongWait = 10 * time.Second

	// pingPeriod is the time between peer pings. Must be less than pongWait.
	pingPeriod = (pongWait * 5) / 10
)

var deadlineForRestart *time.Time

type Parser interface {
	Parse() string
}

type ActionInfo struct {
	ID         string `json:"id"`
	AgentID    string `json:"coreClientID"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Timestamp  int64  `json:"timestamp"`
	ExpireTime int64  `json:"expireTime"`
}

func (a *ActionInfo) Parse() string {
	out, _ := json.Marshal(a)
	return string(out)
}

// GenWSQredoCoreClientFeedURL assembles and returns the Qredo WS client feed URL as a string.
func GenWSQredoCoreClientFeedURL(h *handler) string {
	builder := strings.Builder{}
	builder.WriteString(h.cfg.Base.WsScheme)
	builder.WriteString(h.cfg.Base.QredoAPIDomain)
	builder.WriteString(h.cfg.Base.QredoAPIBasePath)
	builder.WriteString("/coreclient/feed")
	return builder.String()
}

func restartWebSocketHandler(h *handler) {
	h.log.Debug("Handler for restartWebSocketHandler")
	if deadlineForRestart == nil {
		deadlineForRestart = new(time.Time)
		*deadlineForRestart = time.Now()
	} else if deadlineForRestart != nil && time.Since(*deadlineForRestart) >= time.Duration(5*time.Minute) {
		h.log.Error("background job - trying to retry connection failed")
		return
	}
	h.log.Debug("background job - trying to retry connection in next 5 seconds")
	time.Sleep(5 * time.Second)
	go AutoApproveHandler(h)
}

// AutoApproveHandler approves ActionInfo requests read from the Qredo core client-feed websocket.  A connection to
// the websocket is created from information contained in h.  On failure or disconnect the handler is automatically
// restarted.
func AutoApproveHandler(h *handler) {
	h.log.Debug("Handler for AutoApproveHandler")
	agentID := h.core.GetSystemAgentID()
	if len(agentID) == 0 {
		h.log.Info("Agent is not yet configured, skipping Websocket connection for auto-approval")
		return
	}
	url := GenWSQredoCoreClientFeedURL(h)

	// a channel to receive interrupt signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	zkpOnePass, err := h.core.GetAgentZKPOnePass()
	if err != nil {
		h.log.Errorf("cannot get zkp token: ", err)
		return
	}

	headers := http.Header{}
	headers.Set(defs.AuthHeader, hex.EncodeToString(zkpOnePass))

	wsQredoBackedConn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		h.log.Errorf("cannot connect to Websocket feed at Qredo: %v", err)
		go restartWebSocketHandler(h)
		return
	}
	defer wsQredoBackedConn.Close()

	deadlineForRestart = nil // everything is working fine, deadlines should be neutralized
	done := make(chan struct{})

	h.log.Infof("Connected to Qredo websocket feed %s", url)

	// read and process (approve) ActionInfo requests received from the websocket.
	go func() {
		defer func() {
			if err := recover(); err != nil {
				h.log.Errorf("background job - web socket connection panic occurred: %v", err)
			}
			close(done)
			go restartWebSocketHandler(h)
		}()
		for {
			var v Parser = &ActionInfo{}
			if err := wsQredoBackedConn.ReadJSON(v); err != nil {
				h.log.Errorf("error when reading from websocket: %v", err)
			}
			h.log.Infof("background job - incoming message: %v", v.Parse())
			var action ActionInfo
			err = json.Unmarshal([]byte(v.Parse()), &action)
			if action.ExpireTime > time.Now().Unix() {
				go approveActionWithRetry(h, action, 5, 5)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			h.log.Error("interrupt")
			err := wsQredoBackedConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				h.log.Errorf("websocket CloseMessage: %v", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

// approveActionWithRetry attempts to approve the action. It retries every intervalSeconds, giving up after
// maxMinutes have expired.
func approveActionWithRetry(h *handler, action ActionInfo, maxMinutes int, intervalSeconds int) {
	h.log.Debug("Handler for approveActionWithRetry")
	tStart := time.Now()
	baseInc := intervalSeconds
	timeEdge := time.Duration(maxMinutes) * time.Minute
	for {
		err := h.core.ActionApprove(action.ID)
		if err == nil {
			h.log.Infof("Action [%v] approved automatically", action.ID)
			break
		} else {
			h.log.Errorf("Action [%v] approval failed, for [actionID:%v]. Error msg: %v", action.AgentID, action.ID, err)
		}
		if time.Since(tStart) >= timeEdge {
			// Action approval should be abandoned after maxMinutes (e.g. 5 minutes)
			h.log.Warnf("Auto action approve failed [actionID:%v]", action.ID)
			break
		}

		h.log.Warnf("Auto approve action is repeated [actionID:%v] ", action.ID)
		time.Sleep(time.Duration(baseInc) * time.Second)
		baseInc += intervalSeconds
	}
}

func WebSocketFeedHandler(h *handler, w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Handler for WebSocketFeedHandler")

	url := GenWSQredoCoreClientFeedURL(h)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	h.log.Debug(fmt.Sprintf("WebSocketFeedHandler - connecting to %s", url))

	zkpOnePass, err := h.core.GetAgentZKPOnePass()
	if err != nil {
		h.log.Errorf("WebSocketFeedHandler - cannot get zkp token: %v", err)
		return
	}
	headers := http.Header{}
	headers.Set(defs.AuthHeader, hex.EncodeToString(zkpOnePass))

	wsQredoBackedConn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		h.log.Errorf("WebSocketFeedHandler - cannot connect to websocket feed %s: %v", url, err)
		return
	}
	defer wsQredoBackedConn.Close()

	done := make(chan struct{})

	wsPartnerAppUpgrader := websocket.Upgrader{
		ReadBufferSize:  512, // moreless ActionInfo contain 255 B
		WriteBufferSize: 1024,
	}
	wsPartnerAppConn, err := wsPartnerAppUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Errorf("WebSocketFeedHandler - cannot set websocket Partner App Connection: %v", err)
		return
	}
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		wsPartnerAppConn.Close()
		ticker.Stop()
	}()

	wsPartnerAppConn.SetPongHandler(func(message string) error {
		wsPartnerAppConn.SetReadDeadline(time.Now().Add(pongWait))
		return wsPartnerAppConn.WriteControl(websocket.PingMessage, []byte(message), time.Now().Add(writeWait))
	})

	wsPartnerAppConn.SetPingHandler(func(message string) error {
		wsPartnerAppConn.SetWriteDeadline(time.Now().Add(pingPeriod))
		return wsPartnerAppConn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(writeWait))
	})

	h.log.Debugf("WebSocketFeedHandler - Connected to Qredo websocket feed %s", url)
	quitGoRoutine := make(chan bool, 1)
	go func() {
		defer close(done)
	goRoutineLoop:
		for {
			select {
			case quit := <-quitGoRoutine:
				if quit {
					h.log.Debug("WebSocketFeedHandler - terminating reading and writing on websocket conn")
					break goRoutineLoop
				}
			default:
			}

			h.log.Debug("WebSocketFeedHandler - waiting for incoming message")
			var v Parser = &ActionInfo{}

			if err := wsQredoBackedConn.ReadJSON(v); err != nil {
				h.log.Errorf("WebSocketFeedHandler - error when reading from websocket: %v", err)
				break goRoutineLoop
			}

			h.log.Debugf("WebSocketFeedHandler - incoming message: %v", v.Parse())
			err = wsPartnerAppConn.WriteJSON(v)
			if err != nil {
				h.log.Errorf("WebSocketFeedHandler - websocket wsPartnerAppConn WriteJSON contain error: %v", err)
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			wsPartnerAppConn.SetWriteDeadline(time.Now().Add(writeWait))
			err = wsPartnerAppConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(pingPeriod))
			if err != nil {
				h.log.Debug("WebSocketFeedHandler - websocket PingMessage found broken pipe, terminating")
				quitGoRoutine <- true
				return
			}
		case <-done:
			return
		case <-interrupt:
			h.log.Error("WebSocketFeedHandler - interrupt")
			err := wsPartnerAppConn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
			if err != nil {
				h.log.Error("WebSocketFeedHandler - websocket CloseMessage: ", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
