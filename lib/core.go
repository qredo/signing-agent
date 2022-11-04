package lib

import (
	"signing-agent/api"
	"signing-agent/config"
	"signing-agent/util"
)

type SigningAgentClient interface {
	// ClientInit starts the agent registration process
	ClientInit(register *api.QredoRegisterInitRequest, ref, apikey, b64PrivateKey string) (*api.QredoRegisterInitResponse, error)
	// ClientRegister starts the simplified agent registration procedure
	// by generating BLS and EC key pairs and returns the public keys
	ClientRegister(name string) (*api.ClientRegisterResponse, error)
	// ClientRegisterFinish concludes the agent registration process
	ClientRegisterFinish(req *api.ClientRegisterFinishRequest, ref string) (*api.ClientRegisterFinishResponse, error)
	// ClientsList is not currently implemented
	ClientsList() ([]string, error)

	// ActionApprove signs actionID and sends it for approval to the Qredo backend
	ActionApprove(actionID string) error
	// ActionReject sends a rejection to the Qredo backend for actionID
	ActionReject(actionID string) error

	// SetSystemAgentID function to collect agent ID to storage, so the system will default to a single agent ID (AgentID)
	SetSystemAgentID(agetID string) error
	// GetSystemAgentID function to get agent ID that was stored during registration process.
	GetSystemAgentID() string
	// GetAgentZKPOnePass function to generate Zero Knowladge Proof one password (for auth header).
	GetAgentZKPOnePass() ([]byte, error)

	// ReadAction connect to qredo web socket stream by given feed url and return Feed object
	ReadAction(string, ServeCB) *feed
}

type signingAgent struct {
	store *Storage
	cfg   *config.Base
	htc   *util.Client
}

func New(cfg *config.Base, kv util.KVStore) (*signingAgent, error) {
	return &signingAgent{
		cfg:   cfg,
		store: NewStore(kv),
		htc:   util.NewHTTPClient(),
	}, nil
}
