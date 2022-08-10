package lib

import (
	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
)

type CoreClient interface {
	// ClientInit initiate the core client registration process
	ClientInit(register *api.QredoRegisterInitRequest, ref string) (*api.QredoRegisterInitResponse, error)
	// ClientRegister initiates the core client registration procedure
	// by generating BLS and EC key pairs and returns the public keys
	ClientRegister(name string) (*api.ClientRegisterResponse, error)
	// ClientRegisterFinish concludes the core client registration process
	ClientRegisterFinish(req *api.ClientRegisterFinishRequest, ref string) (*api.ClientRegisterFinishResponse, error)
	// ClientsList is not currently implemented
	ClientsList() ([]string, error)

	// ActionApprove signs actionID and sends it for approval to the Qredo backend
	ActionApprove(clientID, actionID string) error
	// ActionReject sends a rejection to the Qredo backend for actionID
	ActionReject(clientID, actionID string) error

	// Sign uses clientID's BLS seed to sign messageHex and returns the signature
	Sign(clientID, messageHex string) (*api.SignResponse, error)
	// Verify verifies a signature provided with VerifyRequest
	Verify(req *api.VerifyRequest) error

	// SetAgentID function to collect Core Client ID to storage, so the system will stand alone per 1 Core Client ID (AgentID)
	SetAgentID(agetID string) error
	// GetAgentID function to get Core Client ID that was storage at registration process.
	GetAgentID() string
}

type coreClient struct {
	store *Storage
	cfg   *config.Base
	htc   *util.Client
}

func New(cfg *config.Base, kv KVStore) (*coreClient, error) {

	return &coreClient{
		cfg:   cfg,
		store: NewStore(kv),
		htc:   util.NewHTTPClient(),
	}, nil
}
