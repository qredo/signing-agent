package api

// swagger:parameters clientRegisterInit
type ClientRegisterRequest struct {
	// in:body
	Name string `json:"name"`
}

// swagger:model clientRegisterResponse
type ClientRegisterResponse struct {
	// in:body
	BLSPublicKey string `json:"bls_public_key"`
	// in:body
	ECPublicKey string `json:"ec_public_key"`
	// in:body
	RefID string `json:"ref_id"`
}

// swagger:parameters clientRegisterFinish
type ClientRegisterFinishRequest struct {
	// in:body
	ID string `json:"id"`
	// in:body
	AccountCode string `json:"account_code"`
	// in:body
	ClientID string `json:"client_id"`
	// in:body
	ClientSecret string `json:"client_secret"`
	// in:body
	IDDoc string `json:"id_doc"`
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
