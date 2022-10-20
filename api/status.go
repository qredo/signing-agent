package api

type WebsocketStatus struct {
	ReadyState       string `json:"ReadyState"`
	RemoteFeedUrl    string `json:"RemoteFeedUrl"`
	LocalFeedUrl     string `json:"LocalFeedUrl"`
	ConnectedClients uint32 `json:"ConnectedClients"`
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

type HealthCheckStatusResponse struct {
	WebsocketStatus WebsocketStatus `json:"WebSocket"`
}
