package lib

import (
	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
)

type CoreClient interface {
	// ClientInit starts the agent registration process
	ClientInit(register *api.QredoRegisterInitRequest, ref string) (*api.QredoRegisterInitResponse, error)
	// ClientRegister starts the simplified agent registration procedure
	// by generating BLS and EC key pairs and returns the public keys
	ClientRegister(name string) (*api.ClientRegisterResponse, error)
	// ClientRegisterFinish concludes the agent registration process
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
	// SetAgentID function to collect agent ID to storage, so the system will default to a single agent ID (AgentID)
	SetAgentID(agetID string) error
	// GetAgentID function to get agent ID that was stored during registration process.
	GetAgentID() string
}

type autoApprover struct {
	store *Storage
	cfg   *config.Base
	htc   *util.Client
}

func New(cfg *config.Base, kv KVStore) (*autoApprover, error) {

	return &autoApprover{
		cfg:   cfg,
		store: NewStore(kv),
		htc:   util.NewHTTPClient(),
	}, nil
}
