package websocket

import (
	"net/http"

	gorilla_ws "github.com/gorilla/websocket"
)

type defaultDialer struct {
	dialer *gorilla_ws.Dialer
}

func NewDefaultDialer() webSocketDialer {
	return &defaultDialer{
		dialer: gorilla_ws.DefaultDialer,
	}
}

func (d *defaultDialer) Dial(url string, requestHeader http.Header) (socketConnection, *http.Response, error) {
	return d.dialer.Dial(url, requestHeader)
}
