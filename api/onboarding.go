package api

import "errors"

// swagger:parameters clientRegisterInit
type ClientRegisterRequest struct {
	// in:body
	Name string `json:"name"`
}

func (r *ClientRegisterRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name")
	}

	return nil
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

func (r *ClientRegisterFinishRequest) Validate() error {
	if r.ID == "" {
		return errors.New("id")
	}
	if r.AccountCode == "" {
		return errors.New("account_code")
	}
	if r.ClientID == "" {
		return errors.New("client_id")
	}
	if r.ClientSecret == "" {
		return errors.New("client_secret")
	}
	if r.IDDoc == "" {
		return errors.New("id_doc")
	}

	return nil
}

// swagger:model clientRegisterFinishResponse
type ClientRegisterFinishResponse struct {
	FeedURL string `json:"feed_url"`
}
