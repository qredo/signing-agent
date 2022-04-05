package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"gitlab.qredo.com/qredo-server/core-client/api"

	"github.com/pkg/errors"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/core-client/lib"
	"gitlab.qredo.com/qredo-server/core-client/util"
)

// Fill in you partner api key and public key pem here
const partner_api_key = ""
const rsa_private_key = ``

type QredoRegisterInitRequest struct {
	Name         string `json:"name"`
	BLSPublicKey string `json:"blsPublicKey"`
	ECPublicKey  string `json:"ecPublicKey"`
}

type QredoRegisterInitResponse struct {
	ID           string `json:"id"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AccountCode  string `json:"accountCode"`
	IDDocument   string `json:"idDoc"`
	Timestamp    int64  `json:"timestamp"`
}

type httpRequest struct {
	url       string
	method    string
	body      []byte
	timestamp string
}

func main() {
	store := util.NewFileStore("examplestore.db")
	cfg := config.Base{
		URL:      "http://127.0.0.1:8007",
		QredoURL: "https://qa-api.qredo.net",
		PIN:      0,
	}

	core, err := lib.New(&cfg, store)
	if err != nil {
		panic(err)
	}

	// initiate on-boarding of new core client
	regInitResponse, err := core.ClientRegister("core-client-test1")
	if err != nil {
		panic(errors.Wrap(err, "core client register init"))
	}

	// send EC and BLS keys to Qredo Partner API, initializing the new core client in the backend
	qredoReq := &QredoRegisterInitRequest{
		Name:         "core-client-test",
		BLSPublicKey: regInitResponse.BLSPublicKey,
		ECPublicKey:  regInitResponse.ECPublicKey,
	}
	body, err := json.Marshal(qredoReq)
	if err != nil {
		panic(err)
	}
	req := &httpRequest{
		timestamp: fmt.Sprintf("%v", time.Now().Unix()),
		method:    "POST",
		url:       fmt.Sprintf("%s/api/v1/p/coreclient/init", cfg.QredoURL),
		body:      body,
	}
	signature := partnerSign(req)
	header := http.Header{}
	header.Add("x-api-key", partner_api_key)
	header.Add("x-sign", signature)
	header.Add("x-timestamp", req.timestamp)
	qredoResponse := &QredoRegisterInitResponse{}
	httpClient := util.NewHTTPClient()
	err = httpClient.Request(req.method, req.url, qredoReq, qredoResponse, header)
	if err != nil {
		panic(errors.Wrap(err, "request to partner api"))
	}

	// finish registration of core client.
	// after this step there will be an entry in examplestore.db for the new core client
	_, err = core.ClientRegisterFinish(&api.ClientRegisterFinishRequest{
		ID:           qredoResponse.ID,
		ClientID:     qredoResponse.ClientID,
		ClientSecret: qredoResponse.ClientSecret,
		AccountCode:  qredoResponse.AccountCode,
		IDDoc:        qredoResponse.IDDocument,
	}, regInitResponse.RefID)
	if err != nil {
		panic(errors.Wrap(err, "client register finish request"))
	}
}

func partnerSign(req *httpRequest) string {
	h := sha256.New()
	h.Write([]byte(req.timestamp))
	h.Write([]byte(req.url))
	h.Write(req.body)
	digest := h.Sum(nil)

	block, _ := pem.Decode([]byte(rsa_private_key))

	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(errors.Wrap(err, "parse RSA key"))
	}

	signature, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, digest)
	if err != nil {
		panic(errors.Wrap(err, "sign request"))
	}

	return base64.RawURLEncoding.EncodeToString(signature)
}
