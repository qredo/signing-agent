package api

// swagger:model WebsocketStatus
type WebsocketStatus struct {
	ReadyState       string `json:"readyState"`
	RemoteFeedUrl    string `json:"remoteFeedURL"`
	LocalFeedUrl     string `json:"localFeedURL"`
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

// swagger:model HealthCheckStatusResponse
type HealthCheckStatusResponse struct {
	WebsocketStatus WebsocketStatus `json:"websocket"`
}
