package websocket

import (
	"net/http"
	"sync"
)

type socketConnection interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type serverConnection interface {
	Connect() bool
	Disconnect()
	Listen(wg *sync.WaitGroup)
	GetFeedUrl() string
	GetReadyState() string
	GetChannel() chan []byte
}

type webSocketDialer interface {
	Dial(url string, requestHeader http.Header) (socketConnection, *http.Response, error)
}
