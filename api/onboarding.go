package api

type ClientRegisterRequest struct {
	Name string `json:"name"`
}

// swagger:model clientRegisterResponse
type ClientRegisterResponse struct {
	BLSPublicKey string `json:"bls_public_key"`
	ECPublicKey  string `json:"ec_public_key"`
	RefID        string `json:"ref_id"`
}

type ClientRegisterFinishRequest struct {
	ID           string `json:"id"`
	AccountCode  string `json:"account_code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	IDDoc        string `json:"id_doc"`
}

// swagger:model clientRegisterFinishResponse
type ClientRegisterFinishResponse struct {
	FeedURL string `json:"feed_url"`
}

type CoreClientServiceRegisterFinishRequest struct {
	ClientID          string `json:"client_id,omitempty"`
	IDDocSignatureHex string `json:"idDocSignatureHex"`
}

type CoreClientServiceRegisterFinishResponse struct {
	Feed string `json:"feed"`
}

type QredoRegisterInitRequest struct {
	Name         string `json:"name"`
	BLSPublicKey string `json:"blsPublicKey"`
	ECPublicKey  string `json:"ecPublicKey"`
}

type QredoRegisterInitResponse struct {
	ID           string `json:"id"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AccountCode  string `json:"accountCode"`
	IDDocument   string `json:"idDoc"`
	Timestamp    int64  `json:"timestamp"`
}

type ClientFullRegisterResponse struct {
	AgentID string `json:"agentID"`
	FeedURL string `json:"feedURL"`
}
