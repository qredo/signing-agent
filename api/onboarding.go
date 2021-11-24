package api

import "errors"

type ClientRegisterRequest struct {
	Name string `json:"name"`
}

func (r *ClientRegisterRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name")
	}

	return nil
}

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

type ClientRegisterFinishResponse struct {
	FeedURL string `json:"feed_url"`
}
