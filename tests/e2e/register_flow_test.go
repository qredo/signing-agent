package e2e_test

import (
	"fmt"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/rest"
	"gitlab.qredo.com/custody-engine/automated-approver/rest/version"
)

const (
	TestQredoAPIDomain      = "play-api.qredo.network"
	TestDataDBStoreFilePath = "../../testdata/test-store.db"
	TestBuildVersion        = "(test-cb12berf)"
	TestBuildType           = "test-dev"
	TestBuildDate           = "Wed 29 Feb 2021 15:28:38 BST"
)

var (
	APIKey           string
	Base64PrivateKey string
)

// Default creates configuration with default values
func createTestConfig() config.Config {
	// can't proceed without an APIKEY or BASE64PKEY
	APIKey = os.Getenv("APIKEY")
	if APIKey == "" {
		log.Fatalf("APIKEY not set in environment")
	}
	Base64PrivateKey = os.Getenv("BASE64PKEY")
	if Base64PrivateKey == "" {
		log.Fatalf("BASE64PKEY not set in environment")
	}

	var cfg config.Config
	cfg.Default()
	cfg.Logging.Level = "debug"
	cfg.Store.FileConfig = TestDataDBStoreFilePath
	cfg.Base.WsScheme = "wss"
	cfg.Base.AutoApprove = true
	cfg.Base.QredoAPIDomain = TestQredoAPIDomain
	return cfg
}

func getTestLog() *zap.SugaredLogger {
	logConfig := zap.NewDevelopmentConfig()
	logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	l, _ := logConfig.Build()
	return l.Sugar()
}

func getTestServer(cfg config.Config) *httptest.Server {
	//func createHttpServer(cfg config.Config) http.Server {
	log := getTestLog()
	ver := version.DefaultVersion()
	ver.BuildVersion = TestBuildVersion
	ver.BuildType = TestBuildType
	ver.BuildDate = TestBuildDate

	router, err := rest.NewQRouter(log, &cfg, ver)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	return httptest.NewServer(router.SetHandlers())
}

// TestRegisterNewSigningAgent will test registering a new agent and that
func TestRegisterNewSigningAgent(t *testing.T) {

	cfg := createTestConfig()
	server := getTestServer(cfg)

	defer func() {
		err := os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
		server.Close()
	}()

	e := httpexpect.New(t, server.URL)

	payload := &api.ClientRegisterRequest{
		Name:             "Agent Test Name",
		APIKey:           string(APIKey),
		Base64PrivateKey: string(Base64PrivateKey),
	}

	// register the client and check against response against expected
	registrationResponse := e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusOK).JSON()

	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("agentId").String().NotEmpty()
	registrationResponse.Object().Value("feedUrl").String().Equal("ws://127.0.0.1:8007/api/v1/client/feed")
	agentID := registrationResponse.Object().Value("agentId").Raw().(string)

	// GET: /client should return the same agentID
	response := e.GET(rest.WrapPathPrefix(rest.PathClientsList)).
		Expect().
		Status(http.StatusOK)
	response.JSON().Array().NotEmpty()
	response.JSON().Array().First().Equal(agentID)
}

// TestRegisterExistingAgentDeny checks attempting to register fails with a known message.
func TestRegisterExistingAgentDeny(t *testing.T) {
	cfg := createTestConfig()
	server := getTestServer(cfg)

	defer func() {
		err := os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
		server.Close()
	}()

	e := httpexpect.New(t, server.URL)

	payload := &api.ClientRegisterRequest{
		Name:             "Agent Test Name",
		APIKey:           string(APIKey),
		Base64PrivateKey: string(Base64PrivateKey),
	}

	// a new client
	registrationResponse := e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusOK).JSON()

	// registering again should result in an error
	registrationResponse = e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusBadRequest).JSON()
	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("Detail").String().Equal("AgentID already exist. You can not set new one.")
}
