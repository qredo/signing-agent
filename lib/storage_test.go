package lib

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/qredo-server/core-client/util"
)

func TestStorage(t *testing.T) {
	kv, err := util.NewFileStore(TestDataDBStoreFilePath)
	defer func() {
		err = os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
	}()

	if err != nil {
		panic(errors.Wrap(err, "file store init"))
	}
	store := NewStore(kv)

	t.Run(
		"Operations on storage - register AgentID",
		func(t *testing.T) {

			val := store.GetAgentID()
			assert.Equal(t, val, "")
			agentID := "5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn"
			store.SetAgentID(agentID)

			val = store.GetAgentID()
			assert.Equal(t, val, agentID)
		})

	t.Run(
		"Operations on storage - registered Client",
		func(t *testing.T) {
			clientID := "5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn"
			client := store.GetClient(clientID)
			assert.Equal(t, client, (*Client)(nil)) // Not found

			client = &Client{
				Name:        "Client Test Name",
				ID:          "7b226964223a22357a5057714c5a61507141614e656e6a797a5779357263614",
				BLSSeed:     []byte("BLSSeed"),
				AccountCode: clientID,
				ZKPID:       []byte("zkpid"),
				ZKPToken:    []byte("zkptoken"),
				Pending:     true,
			}
			err = store.AddClient(clientID, client)
			assert.NoError(t, err)
			takenClient := store.GetClient(clientID)
			assert.Equal(t, takenClient.Name, client.Name)
			assert.Equal(t, takenClient.ID, client.ID)
			assert.Equal(t, takenClient.Pending, false, "Pending mode is not expected.")

			err = store.RemoveClient(clientID)
			assert.NoError(t, err)
			takenClient = store.GetClient(clientID)
			assert.Equal(t, takenClient, (*Client)(nil), "Shouldn't exist anymore.")

			err = store.RemoveClient(clientID)
			assert.Error(t, err, "You can't remove client that doesn't exist.")
		})

	t.Run(
		"Operations on storage - Client with Pending status",
		func(t *testing.T) {

			refID := "47fc1bfa-1aab-4421-aad7-5c42c6e38f1d"
			client := store.GetPending(refID)
			assert.Equal(t, client, (*Client)(nil))

			//add client that is in pending mode
			client = &Client{
				Name:    "Pending Client Test Name",
				Pending: false,
			}
			err = store.AddPending(refID, client)
			assert.NoError(t, err)

			takenClient := store.GetPending(refID)
			assert.Equal(t, takenClient.Name, client.Name)
			assert.Equal(t, takenClient.ID, client.ID)
			assert.Equal(t, takenClient.Pending, true, "Pending mode is expected.")

			err := store.RemovePending(refID)
			assert.NoError(t, err)
		})
}
