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
const partnerAPIKey = "eyJrZXlfaWQiOiJJTnRNNmMwZnFScGJ2ZyIsImtleSI6IkVwSVQxTEFXbTF5TDJhandrUGlmdmprbEMxWktXdHVBLWlVWmJ2RXVUcVUiLCJsYXN0X3VwZGF0ZSI6MCwic2FuZGJveCI6ZmFsc2V9"
const rsaPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA6wER0onSSoNdFYkqcjrYVdRqsT+s2BAcqQXBKKKaADrlbO8F
0qXTjZ4jTqOA9b+BUQhL/fa9fGkBOK+ssGD/f31uH3bdV+8HdHrjzaES3tc0Hv8X
BXoFhO+cEz6xsxWruYxLHEiIY8PxfrNXTPuN128ZTHr6AuxW13xr61P1/CjVjtD5
mZ6ocLuMMPFNzCrn9fWzCEr9YKZhLth7cpgeshX8tyCRBxU0pBUrwO+V6tbAYQu5
WcVuUWbnUnH8GsOyIOrKK3+0N3Z02Oqio4J2n/+N0QlewdAoIULhHphWuSCd3m4x
co/f/6IdT51+ozlv30Ku4nk1nvXyp7MVcRknXQIDAQABAoIBAQDb3WVSSBWSFzMI
igtHUhzCuHiVmpBYmUJnNHYSUYapfnDVqQ8WlITH81LvKPPnd6NCL/QBCE8hzZAR
+/IiFq4UFkLodyoBMiYUaUEmSnPAPzGJanmcaxws0oyASOCyPy0p7ML9FDNeu5z+
QEYGRTfgfhX6QvgTshBRjRve0O/MWJDgOrIWOsPRn9Jgye3GZATEBcnjp+/lgFJn
/LSSwlEuZNUQIYoVKmaHEbrfLlfzljdAdUe5bE/nIfPE0gKT4tWmKvTOIWWxF0Z4
EyWq8t/2p7Uz/R7KTFRKWgh+VTPu8HmDjWBiUKJVGAF9oBV75SaGqkjG8ZGRJ9T1
EAAhwAP1AoGBAPoxaVAvwKqF0mSeDOjA0X+Cn6NDaMgN7LFWMeuqhAmunYCceo7a
YYcUNRAkRQpxRty/brLLV4M+Ii7HDfCuqEPxgWpj5Nv7c/OoMBKCQQzH6eTqemn/
9iSliWfnVlXIKgHmsJ+W+c7QlWMB/8bJGgJr0CyRTrpWvWMG/619abPvAoGBAPB1
aONQoySMe459LQYA+Y7lLqG6B2gbpDsrjVqWi3eGBP93o6pTh2KNpACi+mcPNAyA
2UKIpppiXhKnXYRl5I3VOcGsoen4/pG+T2Ec7pUyXZCixwq82nuiHA3Kpo+yoXFC
iOihpKtMaHUwaaDfPje0c4A6gHEbv+giHC9iON1zAoGBAJFER0WLtG5OLQ7GxfAO
pJVInrAI37noe9mrlmijJO8KN+EI+hAftCjeDsFEjeG2S9K4Q+oELtfBJ8/JO8rX
XlO00dOYFLW1lmmO6fqVLnfhS2jizBjnyV8VzmZJ59L+2YUpELxYyMrQSSynaH9f
HH7zYne+FtwSqPvqgGGXQ9x5AoGBAIWZ5o4uVobPGzNfL23fiskvY0pubwEUIprR
pvdHH/Rn3U0H70KKqHVEl3PXGeO7GcM8r/n8rPyoXPZmUVpntqZra2zFeyzhsKfP
opElnxX8Zuoe1xKLPaVlu8qZ5xN+P58LRcBjV3fpuzwpivbcMtiGhYogdw7hSS40
DY7yNwArAoGAVtikxccT7cY48UtQpp350jPcnym2iTpLzKmBrmwfBwbLPOwKjDrv
iCyzO+LiehYThb1qpiv6sE5Io+qXfDpI8/X+PhRnD2OQJ/Usj1DjVV3uQPe/jHUh
et2Kxa+8d8UuDiJ2PVBIffKrkR9tfTfXdW7Cdh/GNvDSsXdZe4sVzWk=
-----END RSA PRIVATE KEY-----`
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
