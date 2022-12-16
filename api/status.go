package api

// swagger:model WebsocketStatus
type WebsocketStatus struct {
	// The state of the websocket connection with the server
	// enum: OPEN, CLOSED, CONNECTING
	// example: OPEN
	ReadyState string `json:"readyState"`

	// The server websocket URL
	// example: wss://play-api.qredo.network/api/v1/p/coreclient/feed,
	RemoteFeedUrl string `json:"remoteFeedURL"`

	// The local feed websocket URL
	// example: http://localhost:8007/api/v1/client/feed
	LocalFeedUrl string `json:"localFeedURL"`

	// The number of connected feed clients
	// example: 2
	ConnectedClients uint32 `json:"connectedClients"`
}

func NewWebsocketStatus(readyState, remoteFeedUrl, localFeedUrl string, connectedClients int) WebsocketStatus {
	w := WebsocketStatus{
		ReadyState:       readyState,
		RemoteFeedUrl:    remoteFeedUrl,
		LocalFeedUrl:     localFeedUrl,
		ConnectedClients: uint32(connectedClients),
	}
	return w
}

// swagger:model StatusResponse
type HealthCheckStatusResponse struct {
	WebsocketStatus WebsocketStatus `json:"websocket"`
}
