package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
)

const (
	TestDataPrivatePEMFilePath = "../testdata/private.pem"
	TestDataAPIKeyFilePath     = "../testdata/apikey"
	TestDataDBStoreFilePath    = "../testdata/test-store.db"
)

func NewMock(cfg *config.Base, kv KVStore) (*autoApprover, error) {

	return &autoApprover{
		cfg:   cfg,
		store: NewStore(kv),
		htc:   util.NewHTTPMockClient(),
	}, nil
}

func makeCoreHandlerForTests() (*autoApprover, error) {
	var (
		cfg *config.Base
		err error
	)
	cfg = &config.Base{
		URL:                "url",
		PIN:                1234,
		QredoURL:           "https://play-api.qredo.network",
		QredoAPIDomain:     "play-api.qredo.network",
		QredoAPIBasePath:   "/api/v1/p",
		PrivatePEMFilePath: TestDataPrivatePEMFilePath,
		APIKeyFilePath:     TestDataAPIKeyFilePath,
		AutoApprove:        true,
	}

	kv, err := util.NewFileStore(TestDataDBStoreFilePath)
	if err != nil {
		return nil, err
	}
	core, err := NewMock(cfg, kv)
	if err != nil {
		return nil, err
	}
	return core, nil
}

func TestCreateCoreClient(t *testing.T) {
	t.Run(
		"Create a autoApprover",
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
				PrivatePEMFilePath: TestDataPrivatePEMFilePath,
				APIKeyFilePath:     TestDataAPIKeyFilePath,
				AutoApprove:        true,
			}
			cC, err := NewMock(cfg, kv)
			assert.NoError(t, err)
			assert.Equal(t, cC.cfg.PIN, cfg.PIN)
			assert.Equal(t, cC.cfg.PrivatePEMFilePath, cfg.PrivatePEMFilePath)
		})
}
