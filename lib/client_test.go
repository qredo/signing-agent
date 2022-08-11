package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
)

const (
	fixturePathRegisterClientInitResponse = "../testdata/lib/registerClientInitResponse.json"
)

func TestClient(t *testing.T) {
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
	clientName := "Test name client"
	kv, err := util.NewFileStore(TestDataDBStoreFilePath)
	assert.NoError(t, err)
	defer func() {
		err = os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
	}()

	core, err := NewMock(cfg, kv)
	assert.NoError(t, err)
	generatePrivateKey(t, core.cfg.PrivatePEMFilePath)
	err = os.WriteFile(core.cfg.APIKeyFilePath, []byte(""), 0644)
	assert.NoError(t, err)

	var (
		pendingClient          *Client
		registerResponse       *api.ClientRegisterResponse
		initRequest            *api.QredoRegisterInitRequest
		initResponse           *api.QredoRegisterInitResponse
		registerFinishRequest  *api.ClientRegisterFinishRequest
		registerFinishResponse *api.ClientRegisterFinishResponse
	)

	t.Run(
		"Register client - first step",
		func(t *testing.T) {

			registerResponse, err = core.ClientRegister(clientName)
			assert.NoError(t, err)
			assert.NotEmpty(t, registerResponse.BLSPublicKey)
			assert.NotEmpty(t, registerResponse.ECPublicKey)
			assert.NotEmpty(t, registerResponse.RefID)
			pendingClient = core.store.GetPending(registerResponse.RefID)
			assert.Equal(t, true, pendingClient.Pending)
			assert.Equal(t, clientName, pendingClient.Name)
			assert.Empty(t, pendingClient.ID)
		})

	t.Run(
		"Register client - call init ep",
		func(t *testing.T) {
			initRequest = &api.QredoRegisterInitRequest{
				BLSPublicKey: registerResponse.BLSPublicKey,
				ECPublicKey:  registerResponse.ECPublicKey,
				Name:         clientName,
			}

			util.GetDoMockHTTPClientFunc = func(*http.Request) (*http.Response, error) {
				dataFromFixture, err := os.Open(fixturePathRegisterClientInitResponse)
				assert.NoError(t, err)
				body := ioutil.NopCloser(dataFromFixture)

				return &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Body:       body,
				}, nil

			}

			initResponse, err = core.ClientInit(initRequest, registerResponse.RefID)
			assert.NoError(t, err)
			assert.NotEmpty(t, initResponse.AccountCode)

		})

	t.Run(
		"Register client - finish registration",
		func(t *testing.T) {
			util.GetDoMockHTTPClientFunc = func(*http.Request) (*http.Response, error) {

				response := &api.CoreClientServiceRegisterFinishResponse{
					Feed: fmt.Sprintf(
						"ws://%s%s/coreclient/%s/feed",
						core.cfg.QredoAPIDomain,
						core.cfg.QredoAPIBasePath,
						initResponse.AccountCode,
					),
				}

				dataJSON, _ := json.Marshal(response)
				body := ioutil.NopCloser(bytes.NewReader(dataJSON))

				return &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Body:       body,
				}, nil

			}
			registerFinishRequest = &api.ClientRegisterFinishRequest{}
			copier.Copy(&registerFinishRequest, &initResponse)
			registerFinishResponse, err = core.ClientRegisterFinish(registerFinishRequest, registerResponse.RefID)
			assert.NoError(t, err)
			assert.NotEmpty(t, registerFinishResponse.FeedURL)

			// logic verification after registration process
			assert.NotEmpty(t, core.GetAgentID(), "At this stage, we should be able to get AgentID")
			assert.Nil(t, core.store.GetPending(registerResponse.RefID), "At this stage, we shouldn't get pending client")
			registeredClient := core.store.GetClient(initResponse.AccountCode)
			assert.NotNil(t, registeredClient, "At this stage, we should get client")
			assert.False(t, registeredClient.Pending, "Client is not any more at Pending process.")
			assert.NotEmpty(t, registeredClient.ID, "At this stage, client if created properly")
			assert.NotEmpty(t, registeredClient.Name, "At this stage, client if created properly")
			assert.NotEmpty(t, registeredClient.AccountCode, "At this stage, client if created properly")
			assert.NotEmpty(t, registeredClient.BLSSeed, "At this stage, client if created properly")
			assert.NotEmpty(t, registeredClient.ZKPID, "At this stage, client if created properly")
			assert.NotEmpty(t, registeredClient.ZKPToken, "At this stage, client if created properly")
		})

	t.Run(
		"Register client - finish registration fake Agent ID",
		func(t *testing.T) {
			registerFinishRequestFake := &api.ClientRegisterFinishRequest{}
			_, err = core.ClientRegisterFinish(registerFinishRequestFake, "fake RefID")
			assert.Error(t, err)
			_, err = core.ClientRegisterFinish(registerFinishRequestFake, registerResponse.RefID)
			assert.Error(t, err)
			registerFinishRequestFake.ClientID = initResponse.AccountCode
			_, err = core.ClientRegisterFinish(registerFinishRequestFake, registerResponse.RefID)
			assert.Error(t, err)
			registerFinishRequestFake.ClientSecret = initResponse.ClientSecret
			_, err = core.ClientRegisterFinish(registerFinishRequestFake, registerResponse.RefID)
			assert.Error(t, err)
		})

	t.Run(
		"ClientsList",
		func(t *testing.T) {
			var agentsIDList []string
			agentsIDList, err = core.ClientsList()
			assert.NoError(t, err)
			assert.Equal(t, []string{initResponse.AccountCode}, agentsIDList)
		})

	t.Run(
		"Agent - setting and getting",
		func(t *testing.T) {
			agentID := "BbCoiGKwPfc4DYWE6mE2zAEeuEowXLE8sk1Tc9TN8tos"
			core.SetAgentID(agentID)
			assert.Equal(t, core.GetAgentID(), agentID)

			var agentsIDList []string
			agentsIDList, err = core.ClientsList()
			assert.NoError(t, err)
			assert.Equal(t, []string{agentID}, agentsIDList)
		})
}
