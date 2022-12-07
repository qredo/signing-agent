package api

// swagger:ignore
type CoreClientServiceActionMessagesResponse struct {
	Messages []string `json:"messages"`
}

// swagger:ignore
type CoreClientServiceActionApproveRequest struct {
	Signatures []string `json:"signatures"`
	ClientID   string   `json:"client_id,omitempty"`
	ActionID   string   `json:"action_id,omitempty"`
}
