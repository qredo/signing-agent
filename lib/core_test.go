package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
)

const (
	TestDataDBStoreFilePath = "../testdata/test-store.db"
)

func NewMock(cfg *config.Base, kv util.KVStore) (*signingAgent, error) {
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
				kv  util.KVStore
				err error
			)
			cfg = &config.Base{
				PIN:              1234,
				QredoAPIDomain:   "play-api.qredo.network",
				QredoAPIBasePath: "/api/v1/p",
				AutoApprove:      true,
			}

			kv = util.NewFileStore(TestDataDBStoreFilePath)
			err = kv.Init()
			require.Nil(t, err)

			core, err := NewMock(cfg, kv)
			assert.NoError(t, err)
			assert.Equal(t, core.cfg.PIN, cfg.PIN)
		})
}
