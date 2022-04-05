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
const partnerAPIKey = ""
const rsaPrivateKey = ``
const qredoURL = "https://qa-api.qredo.net"

type qredoRegisterInitRequest struct {
	Name         string `json:"name"`
	BLSPublicKey string `json:"blsPublicKey"`
	ECPublicKey  string `json:"ecPublicKey"`
}

type qredoRegisterInitResponse struct {
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
	store, err := util.NewFileStore("examplestore.db")
	if err != nil {
		panic(errors.Wrap(err, "file store init"))
	}
	cfg := config.Base{
		QredoURL: qredoURL,
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
	qredoReq := &qredoRegisterInitRequest{
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
	header.Add("x-api-key", partnerAPIKey)
	header.Add("x-sign", signature)
	header.Add("x-timestamp", req.timestamp)
	qredoResponse := &qredoRegisterInitResponse{}
	httpClient := util.NewHTTPClient()
	err = httpClient.Request(req.method, req.url, qredoReq, qredoResponse, header)
	if err != nil {
		panic(errors.Wrap(err, "request to partner api"))
	}

	// finish registration of core client.
	// after this step there will be an entry in examplestore.db for the new core client
	finishResp, err := core.ClientRegisterFinish(&api.ClientRegisterFinishRequest{
		ID:           qredoResponse.ID,
		ClientID:     qredoResponse.ClientID,
		ClientSecret: qredoResponse.ClientSecret,
		AccountCode:  qredoResponse.AccountCode,
		IDDoc:        qredoResponse.IDDocument,
	}, regInitResponse.RefID)
	if err != nil {
		panic(errors.Wrap(err, "client register finish request"))
	}

	fmt.Printf("created core client\nid: %s\nfeed url: %s\n", qredoResponse.ID, finishResp.FeedURL)
}

func partnerSign(req *httpRequest) string {
	h := sha256.New()
	h.Write([]byte(req.timestamp))
	h.Write([]byte(req.url))
	h.Write(req.body)
	digest := h.Sum(nil)

	block, _ := pem.Decode([]byte(rsaPrivateKey))

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
