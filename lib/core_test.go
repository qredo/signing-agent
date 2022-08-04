package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/core-client/util"
)

func TestCreateCoreClient(t *testing.T) {
	t.Run(
		"Create a coreClient",
		func(t *testing.T) {
			var (
				cfg *config.Base
				kv  KVStore
			)
			cfg = &config.Base{
				URL:                "url",
				PIN:                1234,
				QredoURL:           "https://play-api.qredo.network",
				QredoAPIDomain:     "play-api.qredo.network",
				QredoAPIBasePath:   "/api/v1/p",
				PrivatePEMFilePath: "/path/to/private.pem",
				APIKeyFilePath:     "/path/to/apikey",
				AutoApprove:        true,
			}
			cC, err := New(cfg, kv)
			assert.NoError(t, err)
			assert.Equal(t, cC.cfg.PIN, cfg.PIN)
			assert.Equal(t, cC.cfg.PrivatePEMFilePath, cfg.PrivatePEMFilePath)
			assert.Equal(t, cC.htc, &util.Client{})
		})
}
