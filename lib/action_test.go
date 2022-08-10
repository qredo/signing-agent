package lib

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/core-client/util"
)

const (
	fixturePathClient                  = "../testdata/lib/client.json"
	fixturePathActionApproveGetMessage = "../testdata/lib/actionApproveGetMessage.json"
)

func popMockHttpResponse(alist []*http.Response) *http.Response {
	f := len(alist)
	rv := (alist)[f-1]
	alist = (alist)[:f-1]
	return rv
}

func TestAction(t *testing.T) {
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
	assert.NoError(t, err)
	defer func() {
		err = os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
	}()

	core, err := NewMock(cfg, kv)
	assert.NoError(t, err)
	generatePrivateKey(t, core.cfg.PrivatePEMFilePath)
	defer func() {
		err = os.Remove(TestDataPrivatePEMFilePath)
		assert.NoError(t, err)
	}()
	err = os.WriteFile(core.cfg.APIKeyFilePath, []byte(""), 0644)
	assert.NoError(t, err)
	defer func() {
		err = os.Remove(TestDataAPIKeyFilePath)
		assert.NoError(t, err)
	}()
	var (
		accountCode = "BbCoiGKwPfc4DYWE6mE2zAEeuEowXLE8sk1Tc9TN8tos"
		actionID    = "2D7YA7Ojo3zGRtHP9bw37wF5jq3"
		client      = &Client{}
	)
	data, err := os.ReadFile(fixturePathClient)
	assert.NoError(t, err)
	err = json.Unmarshal(data, client)
	assert.NoError(t, err)
	core.store.AddClient(accountCode, client)

	t.Run(
		"ActionApprove",
		func(t *testing.T) {
			var httpResponseList = []*http.Response{}

			dataFromFixture, err := os.Open(fixturePathActionApproveGetMessage)
			assert.NoError(t, err)
			body := ioutil.NopCloser(dataFromFixture)

			httpResponseMockGetActionMessages := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       body,
			}
			httpResponseMockPutActionApprove := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			}
			httpResponseList = append(httpResponseList, httpResponseMockPutActionApprove)
			httpResponseList = append(httpResponseList, httpResponseMockGetActionMessages)

			util.GetDoMockHTTPClientFunc = func(request *http.Request) (*http.Response, error) {
				// Every time we call mocked server we get response from the stack via popMockHttpResponse func.
				// First we are calling ep to get action message then sign it and PUT to accept ep.
				return popMockHttpResponse(httpResponseList), nil
			}

			err = core.ActionApprove(accountCode, actionID)
			assert.NoError(t, err)
		})

	t.Run(
		"ActionApprove - without first message from Qredo",
		func(t *testing.T) {
			var httpResponseList = []*http.Response{}

			msg := []byte(`{"messages":[]}`)
			body := ioutil.NopCloser(bytes.NewReader(msg))
			httpResponseMockGetActionMessages := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       body,
			}
			httpResponseMockPutActionApprove := &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			}
			httpResponseList = append(httpResponseList, httpResponseMockPutActionApprove)
			httpResponseList = append(httpResponseList, httpResponseMockGetActionMessages)
			util.GetDoMockHTTPClientFunc = func(request *http.Request) (*http.Response, error) {
				return popMockHttpResponse(httpResponseList), nil
			}

			err = core.ActionApprove(accountCode, actionID)
			assert.Error(t, err)
		})

	t.Run(
		"ActionApprove - with first message from Qredo that is wrong",
		func(t *testing.T) {
			var httpResponseList = []*http.Response{}

			msg := []byte(`{"messages":["wrong message that is not a hex"]}`)
			body := ioutil.NopCloser(bytes.NewReader(msg))
			httpResponseMockGetActionMessages := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       body,
			}
			httpResponseMockPutActionApprove := &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			}
			httpResponseList = append(httpResponseList, httpResponseMockPutActionApprove)
			httpResponseList = append(httpResponseList, httpResponseMockGetActionMessages)
			util.GetDoMockHTTPClientFunc = func(request *http.Request) (*http.Response, error) {
				return popMockHttpResponse(httpResponseList), nil
			}

			err = core.ActionApprove(accountCode, actionID)
			assert.Error(t, err)
		})

	t.Run(
		"ActionApprove with fake Agent ID",
		func(t *testing.T) {
			err = core.ActionApprove("fake accountCode", actionID)
			assert.Error(t, err)
		})

	t.Run(
		"ActionReject",
		func(t *testing.T) {
			util.GetDoMockHTTPClientFunc = func(request *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			}
			err = core.ActionReject(accountCode, actionID)
			assert.NoError(t, err)
		})

	t.Run(
		"ActionReject with fake Agent ID",
		func(t *testing.T) {
			err = core.ActionReject("fake accountCode", actionID)
			assert.Error(t, err)
		})
}
