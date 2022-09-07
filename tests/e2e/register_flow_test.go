package e2e_test

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
	"gitlab.qredo.com/custody-engine/automated-approver/rest"
	"gitlab.qredo.com/custody-engine/automated-approver/rest/version"
	"go.uber.org/zap"
)

const (
	TestDataDBStoreFilePath               = "../../testdata/test-store.db"
	TestAPIKeyFilePath                    = "../../testdata/e2e/apikey"
	TestBase64PrivateKeyFilePath          = "../../testdata/e2e/base64privatekey"
	FixturePathRegisterClientInitResponse = "../../testdata/lib/registerClientInitResponse.json"
)

var (
	testAccountCode string
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
	// handler.HandleFunc(uriClientFeed, mockQBClientFeed)
	handler.HandleFunc(uriClientFeed, mockQBClientFeed)

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
		Feed: fmt.Sprintf(
			"ws://e2e-test-server/api/v1/p/coreclient/%s/feed",
			testAccountCode,
		),
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
	// err = ws.WriteMessage(websocket.TextMessage, []byte(`{"id": "Action if needed"}`))
	w.WriteHeader(http.StatusSwitchingProtocols)
}

func TestAutomatedApproverRegisterFlow(t *testing.T) {
	APIKey, err := ioutil.ReadFile(TestAPIKeyFilePath)
	assert.NoError(t, err)
	Base64PrivateKey, err := ioutil.ReadFile(TestBase64PrivateKeyFilePath)
	assert.NoError(t, err)
	cfg := testDefaultConf()
	srvQB := serverMockQB(cfg.Base.QredoAPIBasePath)
	cfg.Base.QredoAPIDomain = strings.ReplaceAll(srvQB.URL, "http://", "")
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
	registrationResponse := servAA.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusOK).JSON()

	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("agentId").NotNull()
	registrationResponse.Object().Value("feedUrl").NotNull()

	testAccountCode := registrationResponse.Object().Value("agentId").Raw().(string)

	response := servAA.GET(rest.WrapPathPrefix(rest.PathClientsList)).
		Expect().
		Status(http.StatusOK)

	response.JSON().Array().NotEmpty()
	response.JSON().Array().First().Equal(testAccountCode)

	// establish websocket connection
	// ws := servAA.GET(rest.WrapPathPrefix(rest.PathClientFeed)).
	// 	WithWebsocketUpgrade().
	// 	Expect().
	// 	//Status(http.StatusSwitchingProtocols).
	// 	Status(http.StatusOK).
	// 	Websocket()
	// defer ws.Disconnect()
	// ws.Subprotocol().Empty()
	// ws.CloseWithText("bye", websocket.CloseNormalClosure)

}
