package lib

import (
	"bytes"
	"encoding/json"
	"errors"

	"gitlab.qredo.com/custody-engine/automated-approver/defs"
)

var agentIDString = "AgentID"

// KVStore is an interface to a simple key-value store used by the core lib
type KVStore interface {
	// Get returns the data for given key. If key is not found, return nil, defs.KVErrNotFound
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
	Del(key string) error
}

type Storage struct {
	kv KVStore
}

func NewStore(store KVStore) *Storage {
	s := &Storage{
		kv: store,
	}

	return s
}

type Agent struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	BLSSeed     []byte `json:"bls_seed"`
	AccountCode string `json:"account_code,omitempty"`
	ZKPID       []byte `json:"zkpid,omitempty"`
	ZKPToken    []byte `json:"zkptoken,omitempty"`
	Pending     bool   `json:"pending"`
}

func (s *Storage) AddPending(ref string, c *Agent) error {
	c.Pending = true
	_, err := s.kv.Get(ref)
	if err != nil && err != defs.KVErrNotFound {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	if err = s.kv.Set(ref, data); err != nil {
		return err
	}

	return nil
}

func (s *Storage) RemovePending(ref string) error {
	d, err := s.kv.Get(ref)
	if err != nil && err != defs.KVErrNotFound {
		return err
	}

	c := &Agent{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return err
	}

	if !c.Pending {
		return errors.New("client not pending")
	}

	return s.kv.Del(ref)
}

func (s *Storage) GetPending(ref string) *Agent {
	d, err := s.kv.Get(ref)
	if err != nil && err != defs.KVErrNotFound {
		return nil
	}

	c := &Agent{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return nil
	}

	if !c.Pending {
		return nil
	}

	return c
}

func (s *Storage) AddClient(id string, c *Agent) error {
	_, err := s.kv.Get(id)
	if err != nil && err != defs.KVErrNotFound {
		return err
	}

	c.Pending = false
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	if err = s.kv.Set(id, data); err != nil {
		return err
	}

	return nil
}

func (s *Storage) RemoveClient(id string) error {
	d, err := s.kv.Get(id)
	if err != nil && err != defs.KVErrNotFound {
		return err
	}

	c := &Agent{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return err
	}

	if c.Pending {
		return errors.New("client pending")
	}

	return s.kv.Del(id)
}

func (s *Storage) GetClient(id string) *Agent {
	d, err := s.kv.Get(id)
	if err != nil && err != defs.KVErrNotFound {
		return nil
	}

	c := &Agent{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return nil
	}

	if c.Pending {
		return nil
	}

	return c
}

func (s *Storage) GetAgentID() string {
	d, err := s.kv.Get(agentIDString)
	if err != nil && err != defs.KVErrNotFound {
		return ""
	}
	return bytes.NewBuffer(d).String()
}

func (s *Storage) SetAgentID(agentID string) error {
	if err := s.kv.Set(agentIDString, []byte(agentID)); err != nil {
		return err
	}
	return nil
}
