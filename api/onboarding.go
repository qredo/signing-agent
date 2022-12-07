package api

import (
	"errors"
	"strings"
)

const (
	maxBase64FieldSize = 16384
	maxStringFieldSize = 256
)

// swagger:model ClientRegisterRequest
type ClientRegisterRequest struct {
	// The name of the agent
	// example: test-agent
	Name string `json:"name"`

	// The api key for the partner api
	// example: eyJrZXkiOiJHM0Fo... (truncated)
	APIKey string `json:"apikey"`

	// The base64 encoded private key pem of which the public key has been registered in the partner api
	// example: LS0tLS1CRUdJTiBS... (truncated)
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

// swagger:ignore
type ClientRegisterResponse struct {
	BLSPublicKey string `json:"bls_public_key"`
	ECPublicKey  string `json:"ec_public_key"`
	RefID        string `json:"ref_id"`
}

// swagger:ignore
type ClientRegisterFinishRequest struct {
	ID           string `json:"id"`
	AccountCode  string `json:"accountCode"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	IDDocument   string `json:"idDoc"`
}

// swagger:ignore
type ClientRegisterFinishResponse struct {
	FeedURL string `json:"feed_url"`
}

// swagger:ignore
type CoreClientServiceRegisterFinishRequest struct {
	ClientID          string `json:"client_id,omitempty"`
	IDDocSignatureHex string `json:"idDocSignatureHex"`
}

// swagger:ignore
type CoreClientServiceRegisterFinishResponse struct {
	Feed string `json:"feed"`
}

// swagger:ignore
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

// swagger:ignore
type QredoRegisterInitResponse struct {
	ID           string `json:"id"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AccountCode  string `json:"accountCode"`
	IDDocument   string `json:"idDoc"`
	Timestamp    int64  `json:"timestamp"`
}

// swagger:model ClientFullRegisterResponse
type ClientFullRegisterResponse struct {
	AgentID string `json:"agentId"`
	FeedURL string `json:"feedUrl"`
}
