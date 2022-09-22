package e2e_test

import (
	"fmt"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/rest"
	"gitlab.qredo.com/custody-engine/automated-approver/rest/version"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	TestDataDBStoreFilePath      = "../../testdata/test-store.db"
	TestAPIKeyFilePath           = "../../testdata/e2e/apikey"
	TestBase64PrivateKeyFilePath = "../../testdata/e2e/base64privatekey"
	TestBuildVersion             = "(test-cb12berf)"
	TestBuildType                = "test-dev"
	TestBuildDate                = "Wed 29 Feb 2021 15:28:38 BST"
)

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// Default creates configuration with default values
func testDefaultConf() config.Config {
	var cfg config.Config
	cfg.Default()
	cfg.Logging.Level = "debug"
	cfg.Base.StoreFile = TestDataDBStoreFilePath
	return cfg
}

func getTestLog() *zap.SugaredLogger {
	logConfig := zap.NewDevelopmentConfig()
	logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	l, _ := logConfig.Build()
	return l.Sugar()
}

func getTestHandlers(cfg config.Config) http.Handler {
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
	return router.SetHandlers()
}

// TestAutomatedApproverRegisterFlow
func TestAutomatedApproverRegisterFlow(t *testing.T) {
	// initialise: default config and keys needed
	APIKey, err := ioutil.ReadFile(TestAPIKeyFilePath)
	assert.NoError(t, err)
	Base64PrivateKey, err := ioutil.ReadFile(TestBase64PrivateKeyFilePath)
	assert.NoError(t, err)
	cfg := testDefaultConf()
	cfg.Base.WsScheme = "ws://"
	cfg.Base.AutoApprove = true
	handlers := getTestHandlers(cfg)

	// local server and expect e2e engine
	server := httptest.NewServer(handlers)
	defer func() {
		err := os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
		server.Close()
	}()
	e := httpexpect.New(t, server.URL)

	// register agent
	payload := &api.ClientRegisterRequest{
		Name:             "Agent Test Name",
		APIKey:           string(APIKey),
		Base64PrivateKey: string(Base64PrivateKey),
	}

	registerClient(e, payload)
}

// registerClient tests the register endpoint (/register). The payload includes the data to be registered.
// Registration fails if the agent is already registered, otherwise the response includes the new agentID
// and feed URL. Both conditions are tested.
func registerClient(e *httpexpect.Expect, payload *api.ClientRegisterRequest) {

	// test registering a new client
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

	// register an existing client (i.e., data from the same payload) should result in an error
	registrationResponse = e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusBadRequest).JSON()
	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("Detail").String().Equal("AgentID already exist. You can not set new one.")
}
