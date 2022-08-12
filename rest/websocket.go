package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	lib "gitlab.qredo.com/custody-engine/automated-approver/lib"
)

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

func GenWSQredoCoreClientFeedURL(h *handler, agentID string, req *lib.Request) {
	builder := strings.Builder{}
	builder.WriteString("wss://")
	builder.WriteString(h.cfg.Base.QredoAPIDomain)
	builder.WriteString(h.cfg.Base.QredoAPIBasePath)
	builder.WriteString("/coreclient/")
	builder.WriteString(agentID)
	builder.WriteString("/feed")
	req.Uri = builder.String()
}

func WebSocketHandler(h *handler, req *lib.Request) {
	h.log.Debug("Handler for WebSocketHandler")
	url := req.Uri

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	h.log.Debug(fmt.Sprintf("connecting to %s", url))

	headers := http.Header{}
	headers.Add("x-api-key", req.ApiKey)
	headers.Add("x-sign", req.Signature)
	headers.Add("x-timestamp", req.Timestamp)

	wsQredoBackedConn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		h.log.Errorf("cannot connect to Websocket feed at Qredo: ", err)
		return
	}
	defer wsQredoBackedConn.Close()

	done := make(chan struct{})

	h.log.Infof("Connected to Qredo websocket feed %s", url)
	go func() {
		defer close(done)
		for {
			var v Parser = &ActionInfo{}
			if err := wsQredoBackedConn.ReadJSON(v); err != nil {
				h.log.Errorf("error when reading from websocket: ", err)
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
				h.log.Error("websocket CloseMessage: ", err)
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

// approveActionWithRetry - Use this function to accept action (transactoin) with the repetition
func approveActionWithRetry(h *handler, action ActionInfo, maxMinutes int, intervalSeconds int) {
	h.log.Debug("\nHandler for approveActionWithRetry")
	tStart := time.Now()
	baseInc := intervalSeconds
	timeEdge := time.Duration(maxMinutes) * time.Minute
	for {
		err := h.core.ActionApprove(action.AgentID, action.ID)
		if err == nil {
			h.log.Infof("[AgentID:%v] Action %v approved automatically", action.AgentID, action.ID)
			break
		} else {
			h.log.Errorf("[AgentID:%v] Action %v approval failed, error msg: %v", action.AgentID, action.ID, err)
		}
		if time.Since(tStart) >= timeEdge {
			// Action Approval should be skiped after maxMinutes is achieved (e.g. 5 minutes)
			h.log.Warnf("Auto action approve failed [AgentID:%v][actionID:%v]", action.AgentID, action.ID)
			break
		}

		h.log.Warnf("Auto approve action is repeated [AgentID:%v][actionID:%v] ", action.AgentID, action.ID)
		time.Sleep(time.Duration(baseInc) * time.Second)
		baseInc += intervalSeconds
	}
}

func WebSocketFeedHandler(h *handler, req *lib.Request, w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Handler for WebSocketFeedHandler")

	url := req.Uri
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	h.log.Debug(fmt.Sprintf("connecting to %s", url))

	headers := http.Header{}
	headers.Add("x-api-key", req.ApiKey)
	headers.Add("x-sign", req.Signature)
	headers.Add("x-timestamp", req.Timestamp)

	wsQredoBackedConn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		h.log.Errorf("cannot connect to websocket feed %s", url, err)
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
		h.log.Errorf("cannot set websocket Partner App Connection: ", err)
		return
	}
	defer wsPartnerAppConn.Close()
	h.log.Debugf("Connected to Qredo websocket feed %s", url)
	go func() {
		defer close(done)
		for {
			var v Parser = &ActionInfo{}
			if err := wsQredoBackedConn.ReadJSON(v); err != nil {
				h.log.Errorf("error when reading from websocket: ", err)
				return
			}
			h.log.Debugf("incoming message: %v", v.Parse())
			err = wsPartnerAppConn.WriteJSON(v)
			if err != nil {
				h.log.Errorf("websocket wsPartnerAppConn WriteJSON contain error: ", err)
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
				h.log.Error("websocket CloseMessage: ", err)
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
