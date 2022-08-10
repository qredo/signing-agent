package lib

import (
	"encoding/json"
	"os"
	"testing"

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

	assert.NoError(t, err)
	cfg := config.Base{
		QredoURL: "https://play-api.qredo.network",
		PIN:      1234,
	}

	core, err := New(&cfg, kv)
	assert.NoError(t, err)

	var (
		clientID            = "BbCoiGKwPfc4DYWE6mE2zAEeuEowXLE8sk1Tc9TN8tos"
		client              = &Client{}
		messageHex   string = "0300228be28434d7362d45a85797da316ae4f3f015fb9c1ecf502d1af8d1f6b6c7e40d28c1b9da2996a095da5829f3e98e"
		signResponse *api.SignResponse
	)
	data, err := os.ReadFile(fixturePathClient)
	assert.NoError(t, err)
	err = json.Unmarshal(data, client)
	assert.NoError(t, err)
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
