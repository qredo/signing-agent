package lib

import (
	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/core-client/rest"
	"go.uber.org/zap"
)

type CoreClient interface {
	ClientRegister(name string) (*api.ClientRegisterResponse, error)
	ClientRegisterFinish(req *api.ClientRegisterFinishRequest, ref string) (*api.ClientRegisterFinishResponse, error)

	ActionApprove(clientID, actionID string) error
	ActionReject(clientID, actionID string) error

	Sign(clientID, messageHex string) (string, error)
	Verify(clientID string, req *api.VerifyRequest) error
}

type coreClient struct {
	log   *zap.SugaredLogger
	store *Storage
	cfg   *config.Config
	htc   *rest.Client
}

func New(log *zap.SugaredLogger, cfg *config.Config, kv KVStore) (*coreClient, error) {

	return &coreClient{
		log:   log,
		cfg:   cfg,
		store: NewStore(kv),
		htc:   rest.NewClient(log),
	}, nil
}
