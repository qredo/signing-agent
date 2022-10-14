package websocket

import (
	"sync"

	"gitlab.qredo.com/custody-engine/automated-approver/defs"
	"go.uber.org/zap"
)

type FeedHub interface {
	Run() bool
	Stop()
	RegisterClient(client *FeedClient)
	UnregisterClient(client *FeedClient)
}

type websocketFeedHub struct {
	serverConn serverConnection
	broadcast  chan []byte
	clients    map[*FeedClient]bool

	register   chan *FeedClient
	unregister chan *FeedClient
	log        *zap.SugaredLogger
	lock       sync.RWMutex
}

func NewFeedHub(server_conn serverConnection, log *zap.SugaredLogger) FeedHub {
	return &websocketFeedHub{
		serverConn: server_conn,
		log:        log,
		clients:    make(map[*FeedClient]bool),
		register:   make(chan *FeedClient),
		unregister: make(chan *FeedClient),
		lock:       sync.RWMutex{},
	}
}

func (w *websocketFeedHub) Run() bool {
	if !w.serverConn.Connect() {
		return false
	}
	var wg sync.WaitGroup
	wg.Add(2)

	w.broadcast = w.serverConn.GetChannel()
	go w.startHub(&wg)
	go w.serverConn.Listen(&wg)

	wg.Wait()
	return true
}

func (w *websocketFeedHub) Stop() {
	if w.serverConn.GetReadyState() == defs.ConnectionState.Open {
		w.serverConn.Disconnect()
	}
}

func (w *websocketFeedHub) RegisterClient(client *FeedClient) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.clients[client] = true
}

func (w *websocketFeedHub) UnregisterClient(client *FeedClient) {
	w.lock.Lock()
	defer w.lock.Unlock()

	close(client.Feed)
	delete(w.clients, client)
}

// TODO - Will be used for the websocket status
// func (w *websocketFeedHub) getExternalFeedClients() int {
// 	w.lock.Lock()
// 	defer w.lock.Unlock()

// 	count := 0
// 	for fc := range w.clients {
// 		if !fc.IsInternal {
// 			count++
// 		}
// 	}

// 	return count
// }

func (w *websocketFeedHub) startHub(wg *sync.WaitGroup) {
	defer func() {
		w.lock.Lock()
		defer w.lock.Unlock()

		for client := range w.clients {
			delete(w.clients, client)
			close(client.Feed)
		}
	}()

	wg.Done()

	for {
		if message, ok := <-w.broadcast; !ok {
			w.log.Debug("feedHub: broadcast channel was closed")
			return
		} else {
			w.lock.Lock()
			w.log.Debugf("feedHub: message received [%v]", string(message))
			for client := range w.clients {
				select {
				case client.Feed <- message:
				default:
					close(client.Feed)
					delete(w.clients, client)
				}
			}
			w.lock.Unlock()
		}
	}
}
