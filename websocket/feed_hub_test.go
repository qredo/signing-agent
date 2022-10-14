package websocket

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/defs"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
	"go.uber.org/goleak"
)

type mockServerConnection struct {
	ConnectCalled       bool
	ListenCalled        bool
	DisconnectCalled    bool
	GetReadyStateCalled bool
	NextConnect         bool
	NextReadyState      string
	RxMessages          chan []byte
}

func (m *mockServerConnection) Connect() bool {
	m.ConnectCalled = true
	return m.NextConnect
}

func (m *mockServerConnection) Disconnect() {
	m.DisconnectCalled = true

}

func (m *mockServerConnection) Listen(wg *sync.WaitGroup) {
	m.ListenCalled = true
	wg.Done()
}

func (m *mockServerConnection) GetFeedUrl() string {
	return ""
}

func (m *mockServerConnection) GetReadyState() string {
	m.GetReadyStateCalled = true
	return m.NextReadyState
}

func (m *mockServerConnection) GetChannel() chan []byte {
	return m.RxMessages
}

func TestFeedHub_Run_fails_to_connect(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockServerConn := &mockServerConnection{}
	feedHub := NewFeedHub(mockServerConn, util.NewTestLogger())

	//Act
	res := feedHub.Run()

	//Assert
	assert.False(t, res)
	assert.True(t, mockServerConn.ConnectCalled)
}

func TestFeedHub_Run_connects_and_listens(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockServerConn := &mockServerConnection{
		NextConnect: true,
		RxMessages:  make(chan []byte),
	}
	feedHub := NewFeedHub(mockServerConn, util.NewTestLogger())

	//Act
	res := feedHub.Run()

	//Assert
	assert.True(t, res)
	assert.True(t, mockServerConn.ConnectCalled)
	assert.True(t, mockServerConn.ListenCalled)

	close(mockServerConn.RxMessages)
}

func TestFeedHub_Stop_not_connected(t *testing.T) {
	//Arrange
	mockServerConn := &mockServerConnection{
		NextConnect: true,
	}
	feedHub := NewFeedHub(mockServerConn, util.NewTestLogger())

	//Act
	feedHub.Stop()

	//Assert
	assert.True(t, mockServerConn.GetReadyStateCalled)
	assert.False(t, mockServerConn.DisconnectCalled)
}

func TestFeedHub_Stop_connected(t *testing.T) {
	//Arrange
	mockServerConn := &mockServerConnection{
		NextConnect:    true,
		NextReadyState: defs.ConnectionState.Open,
	}
	feedHub := NewFeedHub(mockServerConn, util.NewTestLogger())

	//Act
	feedHub.Stop()

	//Assert
	assert.True(t, mockServerConn.GetReadyStateCalled)
	assert.True(t, mockServerConn.DisconnectCalled)
}

func TestFeedHub_Register_Unregister_client(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	feedHub := &websocketFeedHub{
		clients: make(map[*FeedClient]bool),
	}
	client := &FeedClient{
		Feed: make(chan []byte),
	}

	//Act//Assert
	feedHub.RegisterClient(client)
	assert.Equal(t, 1, len(feedHub.clients))

	feedHub.UnregisterClient(client)
	assert.Equal(t, 0, len(feedHub.clients))
}

func TestFeedHub_removes_unlistening_client(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)

	client := NewFeedClient(false)
	feedHub := &websocketFeedHub{
		log:       util.NewTestLogger(),
		clients:   map[*FeedClient]bool{&client: true},
		broadcast: make(chan []byte),
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go feedHub.startHub(&wg)
	wg.Wait()

	//Act
	go func() {
		feedHub.broadcast <- []byte("some message")
	}()

	<-time.After(time.Second)

	//Assert
	assert.Equal(t, 0, len(feedHub.clients))
	close(feedHub.broadcast)
}
