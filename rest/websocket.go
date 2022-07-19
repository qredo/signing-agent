package rest

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsRaw = iota
	wsCoreClient
)

type request struct {
	uri       string
	body      []byte
	apiKey    string
	timestamp string
	signature string
	rsaKey    *rsa.PrivateKey
}

type Parser interface {
	Parse() string
}

type ActionInfo struct {
	ID           string `json:"id"`
	CoreClientID string `json:"coreClientID"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	Timestamp    int64  `json:"timestamp"`
	ExpireTime   int64  `json:"expireTime"`
}

func (a *ActionInfo) Parse() string {
	out, _ := json.Marshal(a)
	return string(out)
}

type AutoApprove struct {
	ID           string `json:"id"`
	CoreClientID string `json:"coreClientID"`
	Status       string `json:"status"`
}

func genTimestamp(req *request) {
	req.timestamp = fmt.Sprintf("%v", time.Now().Unix())
}

func genWSQredoCoreClientFeedURL(coreClientID string, req *request) {
	builder := strings.Builder{}
	builder.WriteString("wss://")
	builder.WriteString(*flagQredoAPIDomain)
	builder.WriteString(*flagQredoAPIBasePath)
	builder.WriteString("/coreclient/")
	builder.WriteString(coreClientID)
	builder.WriteString("/feed")
	req.uri = builder.String()
}

func webSocketHandler(h *handler, req *request, wsType int, w http.ResponseWriter, r *http.Request) {
	// TODO: change fmt to logger
	url := req.uri

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	fmt.Printf("\nconnecting to %s\n", url)

	headers := http.Header{}
	headers.Add("x-api-key", req.apiKey)
	headers.Add("x-sign", req.signature)
	headers.Add("x-timestamp", req.timestamp)

	wsQredoBackedConn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		fmt.Println("cannot connect to Core Client websocket feed at Qredo Backend: ", err)
		return
	}
	defer wsQredoBackedConn.Close()

	done := make(chan struct{})

	fmt.Println("connected to Core Client websocket feed at Qredo Backend")

	wsPartnerAppUpgrader := websocket.Upgrader{
		ReadBufferSize:  512, // moreless ActionInfo contain 255 B
		WriteBufferSize: 1024,
	}
	wsPartnerAppConn, err := wsPartnerAppUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("cannot set websocket Partner App Connection: ", err)
		return
	}
	defer wsPartnerAppConn.Close()

	go func() {
		defer close(done)
		for {
			var v Parser = &ActionInfo{}
			if err := wsQredoBackedConn.ReadJSON(v); err != nil {
				fmt.Println("error when reading from websocket: ", err)
				return
			}
			fmt.Printf("\nincoming message:\n%v\n", v.Parse())
			wsPartnerAppConn.WriteJSON(v)

			var action ActionInfo
			err = json.Unmarshal([]byte(v.Parse()), &action)
			if action.ExpireTime > time.Now().Unix() {
				go approveActionWithRetry(h, action, wsPartnerAppConn, 5, 5)
			}

		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			fmt.Println("interrupt")

			err := wsQredoBackedConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Println("websocket CloseMessage: ", err)
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
func approveActionWithRetry(h *handler, action ActionInfo, wsPartnerAppConn *websocket.Conn, maxMinutes int, intervalSeconds int) {
	fmt.Println("\nHandler for approveActionWithRetry")
	tStart := time.Now()
	baseInc := intervalSeconds
	timeEdge := time.Duration(maxMinutes) * time.Minute
	for {
		err := h.core.ActionApprove(action.CoreClientID, action.ID)
		if err == nil {
			fmt.Printf("\n[CoreClientID:%v] Action %v approved automatically", action.CoreClientID, action.ID)
			wsPartnerAppConn.WriteJSON(AutoApprove{action.ID, action.CoreClientID, "approved"})
			break
		} else {
			fmt.Printf("\n[CoreClientID:%v] Action %v approval failed %v", action.CoreClientID, action.ID, err)
		}

		if time.Since(tStart) >= timeEdge {
			// Action Approval should be skiped after maxMinutes is achieved (e.g. 5 minutes)
			fmt.Printf("\nAuto action approve failed [CoreClientID:%v][actionID:%v] ", action.CoreClientID, action.ID)
			break
		}

		fmt.Printf("\nAuto approve action is repeated [CoreClientID:%v][actionID:%v] ", action.CoreClientID, action.ID)
		time.Sleep(time.Duration(baseInc) * time.Second)
		baseInc += intervalSeconds
	}
}
