package lib

import (
	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/core-client/util"
)

type CoreClient interface {
	// ClientRegister initiates the core client registration procedure
	// by generating BLS and EC key pairs and returns the public keys
	ClientRegister(name string) (*api.ClientRegisterResponse, error)
	// ClientRegisterFinish concludes the core client registration process
	ClientRegisterFinish(req *api.ClientRegisterFinishRequest, ref string) (*api.ClientRegisterFinishResponse, error)
	// ClientList is not currently implemented
	ClientsList() (interface{}, error)

	// ActionApprove signs actionID and sends it for approval to the Qredo backend
	ActionApprove(clientID, actionID string) error
	// ActionReject sends a rejection to the Qredo backend for actionID
	ActionReject(clientID, actionID string) error

	// Sign uses clientID's BLS seed to sign messageHex and returns the signature
	Sign(clientID, messageHex string) (*api.SignResponse, error)
	// Verify verifies a signature provided with VerifyRequest
	Verify(req *api.VerifyRequest) error
}

type coreClient struct {
	store *Storage
	cfg   *config.Config
	htc   *util.Client
}

func New(cfg *config.Config, kv KVStore) (*coreClient, error) {

	return &coreClient{
		cfg:   cfg,
		store: NewStore(kv),
		htc:   util.NewHTTPClient(),
	}, nil
}
