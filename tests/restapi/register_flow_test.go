package restapi_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/defs"
	"gitlab.qredo.com/custody-engine/automated-approver/rest"
	"gitlab.qredo.com/custody-engine/automated-approver/rest/version"
	"go.uber.org/zap"
)

const (
	TestDataDBStoreFilePath               = "../../testdata/test-store.db"
	TestAPIKeyFilePath                    = "../../testdata/e2e/apikey"
	TestBase64PrivateKeyFilePath          = "../../testdata/e2e/base64privatekey"
	FixturePathRegisterClientInitResponse = "../../testdata/lib/registerClientInitResponse.json"
	TestBuildVersion                      = "(test-cb12berf)"
	TestBuildType                         = "test-dev"
	TestBuildDate                         = "Wed 29 Feb 2021 15:28:38 BST"
)

var (
	testAccountCode string
)

// Default creates configuration with default values
func testDefaultConf() config.Config {
	var cfg config.Config
	cfg.Default()
	cfg.Logging.Level = "debug"
	cfg.Store.FileConfig = TestDataDBStoreFilePath
	cfg.Base.HttpScheme = "http"
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

func serverMockQB(QredoAPIBasePath string) *httptest.Server {
	handler := http.NewServeMux()
	uriClientInit := fmt.Sprintf("%s/coreclient/init", QredoAPIBasePath)
	handler.HandleFunc(uriClientInit, mockQBClientInit)

	uriClientInitFinish := fmt.Sprintf("%s/coreclient/finish", QredoAPIBasePath)
	handler.HandleFunc(uriClientInitFinish, mockQBClientInitFinish)

	uriClientFeed := fmt.Sprintf("%s/coreclient/feed", QredoAPIBasePath)
	handler.HandleFunc(uriClientFeed, mockQBClientFeed)

	uriActionFeed := fmt.Sprintf("%s/coreclient/action/{action_id}", QredoAPIBasePath)
	//uriActionFeed := fmt.Sprintf("%s/coreclient/action", QredoAPIBasePath)
	handler.HandleFunc(uriActionFeed, mockQBActionFeed)

	svr := httptest.NewServer(handler)
	return svr
}

func mockQBClientInit(w http.ResponseWriter, r *http.Request) {
	var statusCode int
	statusCode = http.StatusOK

	fixtureFile, err := os.Open(FixturePathRegisterClientInitResponse)
	if err != nil {
		panic("Can't get fixtured ClientInit response file.")
	}

	dataFromFixture, err := io.ReadAll(fixtureFile)
	if err != nil {
		panic("Can't get fixtured ClientInit response content.")
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(dataFromFixture)
}

func mockQBClientRegisterConfirmation(w http.ResponseWriter, r *http.Request) {
	var statusCode int
	statusCode = http.StatusOK

	fixtureFile, err := os.Open(FixturePathRegisterClientInitResponse)
	if err != nil {
		panic("Can't get fixtured ClientInit response file.")
	}

	dataFromFixture, err := io.ReadAll(fixtureFile)
	if err != nil {
		panic("Can't get fixtured ClientInit response content.")
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(dataFromFixture)
}

func mockQBClientInitFinish(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	response := &api.CoreClientServiceRegisterFinishResponse{
		Feed: fmt.Sprintf("ws://e2e-test-server/api/v1/p/coreclient/%s/feed", testAccountCode),
	}

	dataJSON, _ := json.Marshal(response)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(dataJSON)
}

func mockQBClientFeed(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	ws, _ := upgrader.Upgrade(w, r, nil)
	ws.WriteControl(websocket.PingMessage, []byte{}, time.Time{})
	_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"id": "Action if needed"}`))
	w.WriteHeader(http.StatusSwitchingProtocols)
}

// mockQBActionFeed mocks calls to backend to /coreclient/action/{action_id}.
func mockQBActionFeed(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		w.WriteHeader(http.StatusOK)
		var req api.CoreClientServiceActionApproveRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := &api.CoreClientServiceActionMessagesResponse{
		Messages: []string{"aa"},
	}

	dataJSON, _ := json.Marshal(response)
	w.Header().Add("content-type", "test/plain")
	w.Write(dataJSON)
}

// TestRestAPIs  mocks the http server and tests each of the endpoints.
func TestRestAPIs(t *testing.T) {
	// API and Base64Private keys needed for the tests.  API key doesn't need to be real, but the Private
	// key is compared with data in FixturePathRegisterClientInitResponse.
	// defaults if not on the command line.
	APIKey := "not a real key"
	Base64PrivateKey, err := ioutil.ReadFile(TestBase64PrivateKeyFilePath)
	assert.NoError(t, err)

	cfg := testDefaultConf()
	srvQB := serverMockQB(cfg.Base.QredoAPIBasePath)
	cfg.Base.QredoAPIDomain = strings.ReplaceAll(srvQB.URL, "http://", "")
	cfg.AutoApprove = config.AutoApprove{
		Enabled:          true,
		RetryIntervalMax: 300,
		RetryInterval:    5,
	}
	cfg.Websocket = config.WebSocketConf{
		ReconnectTimeOut:  200,
		ReconnectInterval: 15,
		WsScheme:          "ws",
	}
	handlers := getTestHandlers(cfg)
	server := httptest.NewServer(handlers)
	defer func() {
		err := os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
		server.Close()
	}()
	servAA := httpexpect.New(t, server.URL)

	// register agent
	payload := &api.ClientRegisterRequest{
		Name:             "Agent Test Name",
		APIKey:           string(APIKey),
		Base64PrivateKey: string(Base64PrivateKey),
	}

	// API endpoint tests
	registrationTests(servAA, payload)
	healthcheckVersionTests(servAA)
	healthCheckConfigTests(servAA)
	healthcheckStatusTests(servAA, cfg.Base.QredoAPIDomain, defs.ConnectionState.Closed)
	clientActionTests(servAA)
	//TODO - disabled for now, healthcheck end point will be updated
	//	healthcheckStatusTests(servAA, cfg.Base.QredoAPIDomain, defs.ConnectionState.Open)
	websocketTests(servAA)
}

// registrationTests checks the register endpoint (/register). The payload includes the data to be registered.
// Registration fails if the agent is already registered, otherwise the response include the new agentID
// and feed URL. Both conditions are tested.
func registrationTests(e *httpexpect.Expect, payload *api.ClientRegisterRequest) {

	// test registering a new agent
	registrationResponse := e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusOK).JSON()

	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("agentId").String().Equal("5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn")
	registrationResponse.Object().Value("feedUrl").String().Equal("ws://127.0.0.1:8007/api/v1/client/feed")

	// GET: /client should return the same agentID
	testAccountCode = registrationResponse.Object().Value("agentId").Raw().(string)
	response := e.GET(rest.WrapPathPrefix(rest.PathClientsList)).
		Expect().
		Status(http.StatusOK)
	response.JSON().Array().NotEmpty()
	response.JSON().Array().First().Equal(testAccountCode)

	// Register an existing client should result in an error.
	registrationResponse = e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusBadRequest).JSON()
	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("Detail").String().Equal("AgentID already exist. You can not set new one.")
}

// healthCheckVersionTests checks the healthcheck version endpoint (/healthcheck/version).
func healthcheckVersionTests(e *httpexpect.Expect) {
	// GET: healthcheck/version
	hcVersion := e.GET(rest.WrapPathPrefix(rest.PathHealthcheckVersion)).
		Expect().
		Status(http.StatusOK).JSON()
	hcVersion.Object().NotEmpty()
	hcVersion.Object().Keys().ContainsOnly("BuildVersion", "BuildType", "BuildDate")
	hcVersion.Object().ValueEqual("BuildVersion", TestBuildVersion)
	hcVersion.Object().ValueEqual("BuildType", TestBuildType)
	hcVersion.Object().ValueEqual("BuildDate", TestBuildDate)
}

// healthCheckConfigTests checks the healthcheck config endpoint (/healthcheck/config).
func healthCheckConfigTests(e *httpexpect.Expect) {
	hcConfig := e.GET(rest.WrapPathPrefix(rest.PathHealthCheckConfig)).
		Expect().
		Status(http.StatusOK).JSON()

	hcConfig.Object().Keys().Contains("Logging")
	logCfg := hcConfig.Object().Value("Logging").Object()
	logCfg.Value("Format").String().Equal("json")
	logCfg.Value("Level").String().Equal("debug")

	hcConfig.Object().Keys().Contains("Base")
	baseCfg := hcConfig.Object().Value("Base").Object()
	baseCfg.Value("PIN").Equal(0)
	baseCfg.Value("QredoAPIBasePath").String().Equal("/api/v1/p")
	baseCfg.Value("QredoAPIDomain").NotNull()

	hcConfig.Object().Keys().Contains("AutoApprove")
	autoApprove := hcConfig.Object().Value("AutoApprove").Object()
	autoApprove.Value("Enabled").Equal(true)
	autoApprove.Value("RetryIntervalMax").Equal(300)
	autoApprove.Value("RetryInterval").Equal(5)

	hcConfig.Object().Keys().Contains("Websocket")
	websocket := hcConfig.Object().Value("Websocket").Object()
	websocket.Value("ReconnectTimeOut").Equal(200)
	websocket.Value("ReconnectInterval").Equal(15)

	hcConfig.Object().Keys().Contains("Store")
	storeCfg := hcConfig.Object().Value("Store").Object()
	storeCfg.Value("Type").Equal("file")
	storeCfg.Value("FileConfig").Equal(TestDataDBStoreFilePath)

	hcConfig.Object().Keys().Contains("HTTP")
	httpCfg := hcConfig.Object().Value("HTTP").Object()
	httpCfg.Value("Addr").String().Equal("127.0.0.1:8007")
	httpCfg.Value("CORSAllowOrigins").Array().Element(0).String().Equal("*")
	httpCfg.Value("LogAllRequests").Equal(false)

	hcConfig.Object().Keys().Contains("LoadBalancing")
	lbConfig := hcConfig.Object().Value("LoadBalancing").Object()
	lbConfig.Value("Enable").Equal(false)
	lbConfig.Value("OnLockErrorTimeOutMs").Equal(300)
	lbConfig.Value("ActionIDExpirationSec").Equal(6)

	lbConfig.Keys().Contains("RedisConfig")
	redisConfig := lbConfig.Value("RedisConfig").Object()
	redisConfig.Value("Host").Equal("redis")
	redisConfig.Value("Port").Equal(6379)
	redisConfig.Value("Password").Equal("")
	redisConfig.Value("DB").Equal(0)
}

// healthCheckStatusTests checks the healthcheck status endpoint (/healthcheck/status).
func healthcheckStatusTests(e *httpexpect.Expect, host string, wsstatus string) {
	hcStatus := e.GET(rest.WrapPathPrefix(rest.PathHealthCheckStatus)).
		Expect().
		Status(http.StatusOK).JSON()
	hcStatus.Object().NotEmpty()
	hcStatus.Object().Keys().ContainsOnly("WebSocket")
	webSocket := hcStatus.Object().Value("WebSocket").Object()
	webSocket.Value("ReadyState").String().Equal(wsstatus)
	webSocket.Value("RemoteFeedUrl").String().Equal(fmt.Sprintf("ws://%s/api/v1/p/coreclient/feed", host))
	webSocket.Value("LocalFeedUrl").String().Equal("ws://127.0.0.1:8007/api/v1/client/feed")
}

// clientActionTest tests the action endpoints (/client/action/{action_id}.
func clientActionTests(e *httpexpect.Expect) {
	// action approve
	e.PUT(rest.WrapPathPrefix(rest.PathAction)).
		Expect().
		Status(http.StatusOK)

	// action reject
	e.DELETE(rest.WrapPathPrefix(rest.PathAction)).
		Expect().
		Status(http.StatusOK)
}

// websocketTest tests the signing agent websocket (/client/feed)
func websocketTests(e *httpexpect.Expect) {
	ws := e.GET(rest.WrapPathPrefix(rest.PathClientFeed)).
		WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.Subprotocol().Empty()
	ws.CloseWithText("bye", websocket.CloseNormalClosure)
}
