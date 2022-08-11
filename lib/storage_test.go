package lib

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
)

func TestStorage(t *testing.T) {
	kv, err := util.NewFileStore(TestDataDBStoreFilePath)
	defer func() {
		err = os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
	}()
	assert.NoError(t, err)

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
			agentID := "5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn"
			agent := store.GetClient(agentID)
			assert.Equal(t, agent, (*Agent)(nil)) // Not found

			data, err := os.ReadFile(fixturePathAgent)
			assert.NoError(t, err)
			agent = &Agent{}
			err = json.Unmarshal(data, agent)
			assert.NoError(t, err)

			err = store.AddClient(agentID, agent)
			assert.NoError(t, err)
			takenClient := store.GetClient(agentID)
			assert.Equal(t, takenClient.Name, agent.Name)
			assert.Equal(t, takenClient.ID, agent.ID)
			assert.Equal(t, takenClient.Pending, false, "Pending mode is not expected.")

			err = store.RemoveClient(agentID)
			assert.NoError(t, err)
			takenClient = store.GetClient(agentID)
			assert.Equal(t, takenClient, (*Agent)(nil), "Shouldn't exist anymore.")

			err = store.RemoveClient(agentID)
			assert.Error(t, err, "You can't remove agent that doesn't exist.")
		})

	t.Run(
		"Operations on storage - Client with Pending status",
		func(t *testing.T) {

			refID := "47fc1bfa-1aab-4421-aad7-5c42c6e38f1d"
			agent := store.GetPending(refID)
			assert.Equal(t, agent, (*Agent)(nil))

			//add agent that is in pending mode
			agent = &Agent{
				Name:    "Pending Agent Test Name",
				Pending: false,
			}
			err = store.AddPending(refID, agent)
			assert.NoError(t, err)

			takenClient := store.GetPending(refID)
			assert.Equal(t, takenClient.Name, agent.Name)
			assert.Equal(t, takenClient.ID, agent.ID)
			assert.Equal(t, takenClient.Pending, true, "Pending mode is expected.")

			err := store.RemovePending(refID)
			assert.NoError(t, err)
		})
}
