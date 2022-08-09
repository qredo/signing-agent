package lib

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/core-client/util"
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
		pendingClient    *Client
		registerResponse *api.ClientRegisterResponse
		initRequest      *api.QredoRegisterInitRequest
		initResponse     *api.QredoRegisterInitResponse
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

			util.GetDoFunc = func(*http.Request) (*http.Response, error) {
				response := &api.QredoRegisterInitResponse{
					ID:           "5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn",
					ClientID:     "7b226964223a22357a5057714c5a61507141614e656e6a797a577935726361476d345075543162665037344767727a46554a6e222c226375727665223a22424c53333831222c2263726561746564223a313635393631323832317d",
					ClientSecret: "041913ad8a161f5f8a93dd6ad7cb8cd2a80711dedc45f997bf3d21313c14e7e34cb08d6f39189eab30b35c0b8b8f9388ce09414425a9a84ea947212f1830bdef3cf19901566b14e99ba283f9f1895c5f7579d52ec2dd7b59ef8f76b7f7f377e3c3",
					AccountCode:  "5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn",
					IDDocument:   "080212204a224f9732d9ea422dec7f7784b0d643102060bbf8cb08b74c95ce74bfe019734a11717265646f2d636f72652d636c69656e748a0199020a11717265646f2d636f72652d636c69656e7422c00113750e4a911b37bb20eec2e324058d000013783d1a30541a4cd686c3e304c1089c5248d9bb896d178cff805728cdd04d01f5d3b18b440c087e85e26aa258ad66f442ea5698d5b63190f6437e9e4378eaaa146ff2b229d3a2c845b556f2d59804078df017aa84577055a42e2e4188ef3aa895d2193a80e76acfa9cb3d55790eba7db6add587ac75edd6ea8becf52a30210fd5bf806063d47e287f78e1930de89971d75ae4a06cb542dc679f8c8f4856b08415a5639cb1d93e7b9d2a2925e6bff2324104ed61356c3b1fecbd8dc9ed6c259410915bb6f2d1f8f8590b97c7580a753a14a33950b57f41549c8aa6088b9e84ec9fdc89ff002e0131b3514e9a08b674548aa5",
					Timestamp:    time.Now().Unix(),
				}

				dataJSON, _ := json.Marshal(response)
				body := ioutil.NopCloser(bytes.NewReader(dataJSON))

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
}
