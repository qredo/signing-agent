package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
)

const (
	TestDataDBStoreFilePath = "../testdata/test-store.db"
)

func NewMock(cfg *config.Base, kv KVStore) (*signingAgent, error) {

	return &signingAgent{
		cfg:   cfg,
		store: NewStore(kv),
		htc:   util.NewHTTPMockClient(),
	}, nil
}

func TestCreateSigningAgentClient(t *testing.T) {
	t.Run(
		"Create a signingAgent",
		func(t *testing.T) {
			var (
				cfg *config.Base
				kv  KVStore
			)
			cfg = &config.Base{
				PIN:              1234,
				QredoAPIDomain:   "play-api.qredo.network",
				QredoAPIBasePath: "/api/v1/p",
				AutoApprove:      true,
			}
			cC, err := NewMock(cfg, kv)
			assert.NoError(t, err)
			assert.Equal(t, cC.cfg.PIN, cfg.PIN)
		})
}
