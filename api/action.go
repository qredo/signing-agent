package api

type CoreClientServiceActionMessagesResponse struct {
	Messages []string `json:"messages"`
}

type CoreClientServiceActionApproveRequest struct {
	Signatures []string `json:"signatures"`
	ClientID   string   `json:"client_id,omitempty"`
	ActionID   string   `json:"action_id,omitempty"`
}
