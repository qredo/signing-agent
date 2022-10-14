package websocket

import "net/http"

type mockWebSocketDialer struct {
	DialCalled        bool
	LastFeedUrl       string
	LastRequestHeader http.Header
	NextError         error
	NextConn          socketConnection
}

func (m *mockWebSocketDialer) Dial(url string, requestHeader http.Header) (socketConnection, *http.Response, error) {
	m.DialCalled = true
	m.LastFeedUrl = url
	m.LastRequestHeader = requestHeader
	return m.NextConn, nil, m.NextError
}

type mockSocketConnection struct {
	ReadMessageCalled  bool
	WriteMessageCalled bool
	CloseCalled        bool
	LastMessageType    int
	NextMessageType    int
	LastData           []byte
	NextData           []byte
	NextError          error
	read               chan bool
}

func (m *mockSocketConnection) ReadMessage() (messageType int, p []byte, err error) {
	m.ReadMessageCalled = true
	<-m.read

	return m.NextMessageType, m.NextData, m.NextError
}

func (m *mockSocketConnection) WriteMessage(messageType int, data []byte) error {
	m.WriteMessageCalled = true
	m.LastMessageType = messageType
	m.LastData = data
	return m.NextError
}

func (m *mockSocketConnection) Close() error {
	return m.NextError
}
