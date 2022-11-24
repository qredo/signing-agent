package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/rest/version"
)

type mockSourceStats struct {
	GetFeedUrlCalled    bool
	GetReadyStateCalled bool
	NextFeedUrl         string
	NextReadyState      string
}

func (m *mockSourceStats) GetFeedUrl() string {
	m.GetFeedUrlCalled = true
	return m.NextFeedUrl
}

func (m *mockSourceStats) GetReadyState() string {
	m.GetReadyStateCalled = true
	return m.NextReadyState
}

type mockConnectedClients struct {
	GetExternalFeedClientsCalled bool
	NextConnectedClients         int
}

func (m *mockConnectedClients) GetExternalFeedClients() int {
	m.GetExternalFeedClientsCalled = true
	return m.NextConnectedClients
}

func TestHealthCheckHandler_HealthCheckStatus(t *testing.T) {
	//Arrange
	sourceMock := &mockSourceStats{
		NextFeedUrl:    "some feel url",
		NextReadyState: "some ready status",
	}

	connClientsMock := &mockConnectedClients{
		NextConnectedClients: 7,
	}

	handler := NewHealthCheckHandler(sourceMock, nil, nil, connClientsMock, "some local feed")

	req, _ := http.NewRequest("GET", "/path", nil)
	rr := httptest.NewRecorder()

	//Act
	response, err := handler.HealthCheckStatus(nil, rr, req)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	data, _ := json.Marshal(response)
	expected := "{\"WebSocket\":{\"ReadyState\":\"some ready status\",\"RemoteFeedUrl\":\"some feel url\",\"LocalFeedUrl\":\"some local feed\",\"ConnectedClients\":7}}"
	assert.Equal(t, expected, string(data))

	assert.True(t, sourceMock.GetFeedUrlCalled)
	assert.True(t, sourceMock.GetReadyStateCalled)
	assert.True(t, connClientsMock.GetExternalFeedClientsCalled)
}

func TestHealthCheckHandler_HealthCheckVersion(t *testing.T) {
	//Arrange
	version := &version.Version{
		BuildVersion: "some build version",
		BuildType:    "some build type",
		BuildDate:    "some build date",
	}
	handler := NewHealthCheckHandler(nil, version, nil, nil, "")

	req, _ := http.NewRequest("GET", "/path", nil)
	rr := httptest.NewRecorder()

	//Act
	response, err := handler.HealthCheckVersion(nil, rr, req)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	data, _ := json.Marshal(response)
	assert.Equal(t, "{\"BuildVersion\":\"some build version\",\"BuildType\":\"some build type\",\"BuildDate\":\"some build date\"}", string(data))
}

func TestHealthCheckHandler_HealthCheckConfig(t *testing.T) {
	//Arrange
	config := &config.Config{
		Base: config.Base{
			PIN:      25,
			QredoAPI: "some url",
		},
		HTTP: config.HttpSettings{
			Addr: "some address",
		},
	}
	handler := NewHealthCheckHandler(nil, nil, config, nil, "")

	req, _ := http.NewRequest("GET", "/path", nil)
	rr := httptest.NewRecorder()

	//Act
	response, err := handler.HealthCheckConfig(nil, rr, req)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	data, _ := json.Marshal(response)
	expected := "{\"Base\":{\"PIN\":25,\"QredoAPI\":\"some url\"},\"HTTP\":{\"Addr\":\"some address\",\"CORSAllowOrigins\":null,\"LogAllRequests\":false},\"Logging\":{\"Format\":\"\",\"Level\":\"\"},\"LoadBalancing\":{\"Enable\":false,\"OnLockErrorTimeOutMs\":0,\"ActionIDExpirationSec\":0,\"RedisConfig\":{\"Host\":\"\",\"Port\":0,\"Password\":\"\",\"DB\":0}},\"Store\":{\"Type\":\"\",\"FileConfig\":\"\",\"OciConfig\":{\"Compartment\":\"\",\"Vault\":\"\",\"SecretEncryptionKey\":\"\",\"ConfigSecret\":\"\"},\"AwsConfig\":{\"Region\":\"\",\"SecretName\":\"\"}},\"AutoApprove\":{\"Enabled\":false,\"RetryIntervalMax\":0,\"RetryInterval\":0},\"Websocket\":{\"QredoWebsocket\":\"\",\"ReconnectTimeOut\":0,\"ReconnectInterval\":0,\"PingPeriod\":0,\"PongWait\":0,\"WriteWait\":0,\"ReadBufferSize\":0,\"WriteBufferSize\":0}}"
	assert.Equal(t, expected, string(data))
}
