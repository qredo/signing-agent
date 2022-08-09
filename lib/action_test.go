package lib

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/core-client/util"
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
	accountCode := "5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn"
	client := &Client{
		Name:        "Test name client",
		ID:          accountCode,
		BLSSeed:     []byte(`\xe6\u007f\x81ǌ7\xdf\b\xae3N\xd1=\x9e\xffE\xa8a\x8d\xb3\r\x89\xa3Fpj\x9f\x04\xfc;\xed\x05\xbbBv4:\x9aO\x1fd\xd4;lθ\xc3\xc4`),
		AccountCode: accountCode,
		ZKPID:       []byte(`{\"id\":\"5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn\",\"curve\":\"BLS381\",\"created\":1659612821}`),
		ZKPToken:    []byte("\x04\b\x176\xe3\xdd \n\xfbr\xaa\x16\xf5r\x83\xa7\x11 `\x9a\xbc\u007fc\xb1?\x13\xbf\xcd\x12U\a\xbd\xa3\xb1\xea\x17g^tb\"\x10:$\x19`p.\x12\r\x9e\xa2\x9b\x03z|m\r\xc17 \x9f\xf1\xac\xd21M\xbdމ\xbb\xb5\xb7\xf2\bz\x9a\\Q\x0e!\x9e\x911D\x1d\r=\x9c\xa3>-\xfeI\x06X*"),
		Pending:     false,
	}

	core.store.AddClient(accountCode, client)

	t.Run(
		"ActionApprove",
		func(t *testing.T) {
			var httpResponseList = []*http.Response{}
			actionID := "2D7YA7Ojo3zGRtHP9bw37wF5jq3"

			msg := []byte(`{"messages":["08051220b3bc39c21df680e9925bcc6872b06d583545c62f3cca12c6353e8a1d5dbe` +
				`83dc1a20c6e194bd4e2af25d200680555196df700577be01328719085d2bcd0f61efc57b28023297010a06536574746c` +
				`65128c01080212067431203e20301a290a056f776e65721220c6e194bd4e2af25d200680555196df700577be01328719` +
				`085d2bcd0f61efc57b1a250a01731220c6e194bd4e2af25d200680555196df700577be01328719085d2bcd0f61efc57b` +
				`1a260a02743112209d57f713d4b3ae8881ef1609d3a2c74ad3936827464a3ade0ab03fe58a86abf22206736574746c65` +
				`329b010a0c5472616e7366657250757368128a01080412067431203e20301a290a056f776e65721220c6e194bd4e2af2` +
				`5d200680555196df700577be01328719085d2bcd0f61efc57b1a250a01731220c6e194bd4e2af25d200680555196df70` +
				`0577be01328719085d2bcd0f61efc57b1a260a02743112209d57f713d4b3ae8881ef1609d3a2c74ad3936827464a3ade` +
				`0ab03fe58a86abf22204707573685220d827677997c9a2ae5210d5ba25f1da6fdfb3fd70e2bcb79cd5f70dacba642711` +
				`5a1273746f636b20617072696c207370696465727a3808042234108032222a3078613242613739423935336343353161` +
				`3738314137363230633233653041443238616262373639633840b0a1c001"]}`)
			body := ioutil.NopCloser(bytes.NewReader(msg))
			httpResponseMockGetActionMessages := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       body,
			}
			// tutaj powinien byc msg ze status approved
			httpResponseMockPutActionApprove := &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       nil,
			}
			httpResponseList = append(httpResponseList, httpResponseMockPutActionApprove)
			httpResponseList = append(httpResponseList, httpResponseMockGetActionMessages)
			util.GetDoMockHTTPClientFunc = func(request *http.Request) (*http.Response, error) {
				return popMockHttpResponse(httpResponseList), nil
			}

			err = core.ActionApprove(accountCode, actionID)
			assert.NoError(t, err)
		})
}
