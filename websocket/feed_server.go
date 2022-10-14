package websocket

import (
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/defs"
	"gitlab.qredo.com/custody-engine/automated-approver/lib"
	"go.uber.org/zap"
)

type websocketServerConn struct {
	dialer               webSocketDialer
	conn                 socketConnection
	log                  *zap.SugaredLogger
	core                 lib.SigningAgentClient
	feedUrl              string
	shouldReconnect      bool
	readyState           string
	reconnectIntervalMax time.Duration
	reconnectInterval    time.Duration
	rxMessages           chan []byte
}

func NewServerConnection(dialer webSocketDialer, feedUrl string, log *zap.SugaredLogger, core lib.SigningAgentClient, config *config.WebSocketConf) serverConnection {
	return &websocketServerConn{
		log:                  log,
		core:                 core,
		feedUrl:              feedUrl,
		dialer:               dialer,
		shouldReconnect:      true,
		readyState:           defs.ConnectionState.Closed,
		reconnectIntervalMax: time.Duration(config.ReconnectTimeOut) * time.Second,
		reconnectInterval:    time.Duration(config.ReconnectInterval) * time.Second,
		rxMessages:           make(chan []byte),
	}
}

func (w *websocketServerConn) Connect() bool {
	w.readyState = defs.ConnectionState.Connecting

	startTime := time.Now()
	for time.Since(startTime) < w.reconnectIntervalMax {
		if err := w.dial(); err == nil {
			w.log.Infof("connected to feed %v", w.feedUrl)
			return true
		} else {
			w.log.Errorf("cannot connect to feed: %v, retry connection in %v", err, w.reconnectInterval)
			time.Sleep(w.reconnectInterval)
		}
	}

	w.readyState = defs.ConnectionState.Closed
	return false
}

func (w *websocketServerConn) Listen(wg *sync.WaitGroup) {
	defer func() {
		w.conn.Close()
		close(w.rxMessages)
	}()

	wg.Done()

	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			//closed on request
			if !w.shouldReconnect {
				return
			}
			//either connection issue or issue reading the message
			w.log.Errorf("unexpected connection error: %v", err)
			if !w.Connect() {
				return
			}
		} else {
			w.rxMessages <- message
		}
	}

}

func (w *websocketServerConn) Disconnect() {
	w.log.Infof("disconnecting from feed %v", w.feedUrl)
	if err := w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		w.log.Errorf("error on send CloseMessage, error: %v", err)
	}
	w.shouldReconnect = false
	w.readyState = defs.ConnectionState.Closed
}

func (w *websocketServerConn) GetFeedUrl() string {
	return w.feedUrl
}

func (w *websocketServerConn) GetReadyState() string {
	return w.readyState
}

func (w *websocketServerConn) GetChannel() chan []byte {
	return w.rxMessages
}

func (w *websocketServerConn) dial() error {
	headers, err := w.genConnectHeaders()
	if err != nil {
		w.log.Errorf("failed to generate dial headers: %v", err)
		return err
	}

	conn, _, err := w.dialer.Dial(w.feedUrl, headers)
	if err == nil {
		w.conn = conn
		w.readyState = defs.ConnectionState.Open
		return nil
	}

	return err
}

func (w *websocketServerConn) genConnectHeaders() (http.Header, error) {
	zkpOnePass, err := w.core.GetAgentZKPOnePass()
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set(defs.AuthHeader, hex.EncodeToString(zkpOnePass))
	return headers, nil
}
