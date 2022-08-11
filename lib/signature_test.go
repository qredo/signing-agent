package lib

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
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
		agentID             = "BbCoiGKwPfc4DYWE6mE2zAEeuEowXLE8sk1Tc9TN8tos"
		agent               = &Agent{}
		messageHex   string = "0300228be28434d7362d45a85797da316ae4f3f015fb9c1ecf502d1af8d1f6b6c7e40d28c1b9da2996a095da5829f3e98e"
		signResponse *api.SignResponse
	)
	data, err := os.ReadFile(fixturePathAgent)
	assert.NoError(t, err)
	err = json.Unmarshal(data, agent)
	assert.NoError(t, err)
	core.store.AddClient(agentID, agent)
	t.Run(
		"Sign the message",
		func(t *testing.T) {

			signResponse, err = core.Sign(agentID, messageHex)
			assert.NoError(t, err)
			assert.Equal(t, agentID, signResponse.SignerID)
			assert.NotNil(t, signResponse.SignatureHex)
		})
	t.Run(
		"Signing the message will fail - wrong SignatureHex",
		func(t *testing.T) {
			_, err := core.Sign(agentID, "this is not hex")
			assert.Error(t, err)
		})
	t.Run(
		"Signing the message will fail - agent not found",
		func(t *testing.T) {
			_, err := core.Sign("agent not found", messageHex)
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
	core.store.AddClient(agentID, agent)
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
