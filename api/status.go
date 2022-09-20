package api

type WebsocketStatus struct {
	ReadyState       string `json:"ReadyState"`
	RemoteFeedUrl    string `json:"RemoteFeedUrl"`
	LocalFeedUrl     string `json:"LocalFeedUrl"`
	ConnectedClients uint32 `json:"ConnectedClients,omitempty"`
}

func NewWebsocketStatus(readyState, remoteFeedUrl, localFeedUrl string) WebsocketStatus {
	w := WebsocketStatus{
		ReadyState:    readyState,
		RemoteFeedUrl: remoteFeedUrl,
		LocalFeedUrl:  localFeedUrl,
	}
	return w
}

type HealthCheckStatusResponse struct {
	WebsocketStatus WebsocketStatus `json:"WebSocket"`
}
