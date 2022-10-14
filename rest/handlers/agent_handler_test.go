package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/autoapprover"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/defs"
	"gitlab.qredo.com/custody-engine/automated-approver/lib"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
	"gitlab.qredo.com/custody-engine/automated-approver/websocket"
	"go.uber.org/goleak"
)

type mockFeedHub struct {
	NextRun                bool
	RunCalled              bool
	RegisterClientCalled   bool
	UnregisterClientCalled bool
	StopCalled             bool
	LastRegisteredClient   *websocket.FeedClient
	LastUnregisteredClient *websocket.FeedClient
}

func (m *mockFeedHub) Run() bool {
	m.RunCalled = true
	return m.NextRun
}

func (m *mockFeedHub) Stop() {
	m.StopCalled = true
}

func (m *mockFeedHub) RegisterClient(client *websocket.FeedClient) {
	m.RegisterClientCalled = true
	m.LastRegisteredClient = client
}

func (m *mockFeedHub) UnregisterClient(client *websocket.FeedClient) {
	m.UnregisterClientCalled = true
	m.LastUnregisteredClient = client
}

var testLog = util.NewTestLogger()

func NewTestRequest() *http.Request {
	test_req, _ := http.NewRequest("POST", "/path", bytes.NewReader([]byte(`
	{
		"Name":"test name",
		"APIKey":"test api key",
		"Base64PrivateKey":"test 64 private key"
	}`)))
	return test_req
}

var testClientRegisterResponse = &api.ClientRegisterResponse{
	ECPublicKey:  "ec",
	BLSPublicKey: "bls",
	RefID:        "refId",
}

var testRegisterInitResponse = &api.QredoRegisterInitResponse{
	ID:           "some id",
	ClientID:     "client id",
	ClientSecret: "client secret",
	AccountCode:  "account code",
	IDDocument:   "iddocument",
	Timestamp:    15456465,
}

func TestSigningAgentHandler_RegisterAgent_already_registered(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextAgentID: "some agent id",
	}
	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil)

	rr := httptest.NewRecorder()

	//Act
	response, err := handler.RegisterAgent(nil, rr, nil)

	//Assert
	assert.Nil(t, response)
	assert.True(t, mock_core.GetSystemAgentIDCalled)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.NotNil(t, err)

	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "AgentID already exist. You can not set new one.", detail)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.False(t, mock_core.ClientRegisterCalled)
	assert.False(t, mock_core.ClientInitCalled)
}

func TestSigningAgentHandler_RegisterAgent_fails_to_decode_request(t *testing.T) {
	//Arrange
	mock_core := lib.NewMockSigningAgentClient("")

	var lastDecoded *http.Request
	decode := func(i interface{}, r *http.Request) error {
		lastDecoded = r
		return defs.ErrBadRequest().WithDetail("some decode error")
	}

	handler := &SigningAgentHandler{
		feedHub: &mockFeedHub{},
		core:    mock_core,
		log:     testLog,
		decode:  decode,
	}

	req, _ := http.NewRequest("POST", "/path", nil)
	rr := httptest.NewRecorder()

	//Act
	response, err := handler.RegisterAgent(nil, rr, req)

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "some decode error", detail)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.True(t, mock_core.GetSystemAgentIDCalled)
	assert.False(t, mock_core.ClientRegisterCalled)
	assert.False(t, mock_core.ClientInitCalled)
	assert.NotNil(t, lastDecoded)
	assert.Equal(t, lastDecoded, req)
}

func TestSigningAgentHandler_RegisterAgent_doesnt_validate_request(t *testing.T) {
	//Arrange
	mock_core := lib.NewMockSigningAgentClient("")

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil)

	req, _ := http.NewRequest("POST", "/path", bytes.NewReader([]byte(`
	{
		"APIKey":"key",
		"Base64PrivateKey":"key"
	}`)))

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), req)

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "name", detail)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestSigningAgentHandler_RegisterAgent_fails_to_register_client(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextRegisterError: errors.New("some error"),
	}

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil)

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, "some error", err.Error())
	assert.True(t, mock_core.ClientRegisterCalled)
	assert.False(t, mock_core.ClientInitCalled)
	assert.Equal(t, "test name", mock_core.LastRegisteredName)
}

func TestSigningAgentHandler_RegisterAgent_fails_to_init_registration(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextClientInitError:        errors.New("some error"),
		NextClientRegisterResponse: testClientRegisterResponse,
	}

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil)

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.True(t, mock_core.ClientInitCalled)
	assert.Equal(t, "some error", err.Error())
	assert.Equal(t, "refId", mock_core.LastRef)
	assert.Equal(t, "test api key", mock_core.LastApiKey)
	assert.Equal(t, "test 64 private key", mock_core.Last64PrivateKey)
	assert.Equal(t, "bls", mock_core.LastRegisterRequest.BLSPublicKey)
	assert.Equal(t, "ec", mock_core.LastRegisterRequest.ECPublicKey)
	assert.Equal(t, "test name", mock_core.LastRegisterRequest.Name)
}

func TestSigningAgentHandler_RegisterAgent_fails_to_finish_registration(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextRegisterFinishError:    errors.New("some error"),
		NextClientRegisterResponse: testClientRegisterResponse,
		NextRegisterInitResponse:   testRegisterInitResponse,
	}

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil)

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, "some error", err.Error())
	assert.True(t, mock_core.ClientRegisterFinishCalled)
	assert.Equal(t, "refId", mock_core.LastRef)

	assert.Equal(t, "account code", mock_core.LastRegisterFinishRequest.AccountCode)
	assert.Equal(t, "client id", mock_core.LastRegisterFinishRequest.ClientID)
	assert.Equal(t, "client secret", mock_core.LastRegisterFinishRequest.ClientSecret)
	assert.Equal(t, "some id", mock_core.LastRegisterFinishRequest.ID)
	assert.Equal(t, "iddocument", mock_core.LastRegisterFinishRequest.IDDocument)
}

func TestSigningAgentHandler_RegisterAgent_returns_response(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextClientRegisterResponse: testClientRegisterResponse,
		NextRegisterInitResponse:   testRegisterInitResponse,
		NextRegisterFinishResponse: &api.ClientRegisterFinishResponse{},
	}

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{
			Addr: "some address",
		}}, nil)

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, mock_core.ClientRegisterFinishCalled)

	res, ok := response.(api.ClientFullRegisterResponse)
	assert.True(t, ok)
	assert.NotNil(t, res)
	assert.Equal(t, "account code", res.AgentID)
	assert.Equal(t, "ws://some address/api/v1/client/feed", res.FeedURL)
}

func TestSigningAgentHandler_StartAgent_runs_feedHub(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{}
	mockCore := lib.NewMockSigningAgentClient("valid_agentID")
	handler := NewSigningAgentHandler(mockFeedHub, mockCore, testLog,
		&config.Config{
			HTTP: config.HttpSettings{}}, nil)

	//Act
	handler.StartAgent()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestSigningAgentHandler_StartAgent_doesnt_start_autoApproval(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}
	mockCore := lib.NewMockSigningAgentClient("valid_agentID")
	config := &config.Config{
		HTTP:        config.HttpSettings{},
		AutoApprove: config.AutoApprove{},
	}
	handler := NewSigningAgentHandler(mockFeedHub, mockCore, testLog, config, nil)

	//Act
	handler.StartAgent()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestSigningAgentHandler_StartAgent_registers_auto_approval(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}
	mockCore := lib.NewMockSigningAgentClient("valid_agentID")

	handler := NewSigningAgentHandler(mockFeedHub, mockCore, testLog, &config.Config{
		AutoApprove: config.AutoApprove{
			Enabled: true,
		},
	}, autoapprover.NewAutoApproval(mockCore, testLog, &config.Config{}, nil, nil))

	//Act
	handler.StartAgent()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.True(t, mockFeedHub.RegisterClientCalled)

	auto_approver := mockFeedHub.LastRegisteredClient
	assert.NotNil(t, auto_approver)
	assert.True(t, auto_approver.IsInternal)
	close(auto_approver.Feed)
}

func TestSigningAgentHandler_StopAgent(t *testing.T) {
	//Arrange
	mockFeedHub := &mockFeedHub{}
	handler := NewSigningAgentHandler(mockFeedHub, nil, util.NewTestLogger(), &config.Config{
		HTTP: config.HttpSettings{}}, &autoapprover.AutoApproval{})

	//Act
	handler.StopAgent()

	//Assert
	assert.True(t, mockFeedHub.StopCalled)
}
