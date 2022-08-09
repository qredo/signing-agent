package lib

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/core-client/config"
	"gitlab.qredo.com/qredo-server/core-client/util"
)

func TestSignature(t *testing.T) {
	kv, err := util.NewFileStore(TestDataDBStoreFilePath)
	defer func() {
		err = os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
	}()

	if err != nil {
		panic(errors.Wrap(err, "file store init"))
	}
	cfg := config.Base{
		QredoURL: "https://play-api.qredo.network",
		PIN:      1234,
	}

	core, err := New(&cfg, kv)
	if err != nil {
		panic(err)
	}
	clientID := "BbCoiGKwPfc4DYWE6mE2zAEeuEowXLE8sk1Tc9TN8tos"

	client := &Client{
		Name:        "Client Test Name",
		ID:          "7b226964223a22357a5057714c5a61507141614e656e6a797a5779357263614",
		BLSSeed:     []byte("9fe24e8115b084c58dbd90d2e0e06d82d989b10d6c670a287908f4fb18bfc2fd5e56852b1f9175560192e30be4a04175"),
		AccountCode: clientID,
		ZKPID:       []byte("7b226964223a224262436f69474b775066633444595745366d45327a41456575456f77584c4538736b31546339544e38746f73222c226375727665223a22424c53333831222c2263726561746564223a313635383134313637337d"),
		ZKPToken:    []byte("0400f84188df4b2786e59daca3162eea1387d0a9898a41d10ad0e7c848824a047af182c7bcb8405dba7197480255e7adfb04c8d604fda55fadbda4383baf2620db2d90ca051a609652db0a2c652dca4eb0532b502a30f1d8577ae69e49ee0b67b8"),
		Pending:     true,
	}
	var (
		messageHex   string = "0300228be28434d7362d45a85797da316ae4f3f015fb9c1ecf502d1af8d1f6b6c7e40d28c1b9da2996a095da5829f3e98e"
		signResponse *api.SignResponse
	)

	core.store.AddClient(clientID, client)
	t.Run(
		"Sign the message",
		func(t *testing.T) {

			signResponse, err = core.Sign(clientID, messageHex)
			assert.NoError(t, err)
			assert.Equal(t, clientID, signResponse.SignerID)
			assert.NotNil(t, signResponse.SignatureHex)
		})
	t.Run(
		"Signing the message will fail - wrong SignatureHex",
		func(t *testing.T) {
			_, err := core.Sign(clientID, "this is not hex")
			assert.Error(t, err)
		})
	t.Run(
		"Signing the message will fail - client not found",
		func(t *testing.T) {
			_, err := core.Sign("client not found", messageHex)
			assert.Error(t, err)
		})
	t.Run(
		"Verify the message - OK",
		func(t *testing.T) {
			var req = &api.VerifyRequest{
				MessageHashHex: messageHex,
				SignatureHex:   signResponse.SignatureHex,
				SignerID:       signResponse.SignerID,
			}
			err = core.Verify(req)
			assert.NoError(t, err)
		})
	core.store.AddClient(clientID, client)
	t.Run(
		"Verifying the message will fail - wrong SignatureHex",
		func(t *testing.T) {
			var req = &api.VerifyRequest{
				MessageHashHex: messageHex,
				SignatureHex:   "string - not hex",
				SignerID:       signResponse.SignerID,
			}
			err = core.Verify(req)
			assert.Error(t, err, "We expected to have error.")
		})
	t.Run(
		"Verifying the message will fail - too short SignatureHex",
		func(t *testing.T) {
			var req = &api.VerifyRequest{
				MessageHashHex: messageHex,
				SignatureHex:   "0300228be28434d7",
				SignerID:       signResponse.SignerID,
			}
			err = core.Verify(req)
			assert.Error(t, err, "We expected to have error.")
		})
	t.Run(
		"Verifying the message will fail - client not found",
		func(t *testing.T) {
			var req = &api.VerifyRequest{
				MessageHashHex: messageHex,
				SignatureHex:   signResponse.SignatureHex,
				SignerID:       "client not found",
			}
			err = core.Verify(req)
			assert.Error(t, err, "We expected to have error.")
			assert.Equal(t, err.Error(), "Not Found: get signer client not found")
		})

}
