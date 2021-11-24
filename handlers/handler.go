package handlers

import (
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/core-client/qhttp/client"
	"go.uber.org/zap"
)

type Handler struct {
	log   *zap.SugaredLogger
	store *Storage
	cfg   *config.Config
	htc   *client.Client
}

func New(log *zap.SugaredLogger, cfg *config.Config) (*Handler, error) {
	storage, err := NewStore(cfg.StoreFile)
	if err != nil {
		return nil, err
	}
	return &Handler{
		log:   log,
		cfg:   cfg,
		store: storage,
		htc:   client.NewClient(log),
	}, nil
}
