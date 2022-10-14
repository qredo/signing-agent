package api

import (
	"errors"
	"strings"
)

const (
	maxBase64FieldSize = 16384
	maxStringFieldSize = 256
)

type ClientRegisterRequest struct {
	Name             string `json:"name"`
	APIKey           string `json:"apikey"`
	Base64PrivateKey string `json:"base64privatekey"`
}

func (r *ClientRegisterRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.APIKey = strings.TrimSpace(r.APIKey)
	r.Base64PrivateKey = strings.TrimSpace(r.Base64PrivateKey)

	switch {
	case r.Name == "" || len(r.Name) > maxStringFieldSize:
		return errors.New("name")
	case r.APIKey == "" || len(r.APIKey) > maxStringFieldSize:
		return errors.New("apikey")
	case r.Base64PrivateKey == "" || len(r.Base64PrivateKey) > maxBase64FieldSize:
		return errors.New("base64privatekey")
	default:
		return nil
	}
}

// swagger:model clientRegisterResponse
type ClientRegisterResponse struct {
	BLSPublicKey string `json:"bls_public_key"`
	ECPublicKey  string `json:"ec_public_key"`
	RefID        string `json:"ref_id"`
}

type ClientRegisterFinishRequest struct {
	ID           string `json:"id"`
	AccountCode  string `json:"accountCode"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	IDDocument   string `json:"idDoc"`
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

func NewQredoRegisterInitRequest(name, blsPublicKey, ecPublicKey string) *QredoRegisterInitRequest {
	return &QredoRegisterInitRequest{
		Name:         name,
		BLSPublicKey: blsPublicKey,
		ECPublicKey:  ecPublicKey,
	}
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
	AgentID string `json:"agentId"`
	FeedURL string `json:"feedUrl"`
}
