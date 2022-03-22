package lib

import (
	"encoding/json"
	"errors"

	"gitlab.qredo.com/qredo-server/core-client/defs"
)

type KVStore interface {
	// Get returns the data for given key. If key is not found, return nil, defs.ErrNotFound
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

type Client struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	BLSSeed     []byte `json:"bls_seed"`
	AccountCode string `json:"account_code,omitempty"`
	ZKPID       []byte `json:"zkpid,omitempty"`
	ZKPToken    []byte `json:"zkptoken,omitempty"`
	Pending     bool   `json:"pending"`
}

func (s *Storage) AddPending(ref string, c *Client) error {
	c.Pending = true
	_, err := s.kv.Get(ref)
	if err != nil && err != defs.ErrNotFound {
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
	if err != nil {
		if err != defs.ErrNotFound {
			return nil
		}
		return err
	}

	c := &Client{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return err
	}

	if !c.Pending {
		return errors.New("client not pending")
	}

	return s.kv.Del(ref)
}

func (s *Storage) GetPending(ref string) *Client {
	d, err := s.kv.Get(ref)
	if err != nil {
		if err != defs.ErrNotFound {
			return nil
		}
		return nil
	}

	c := &Client{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return nil
	}

	if !c.Pending {
		return nil
	}

	return nil
}

func (s *Storage) AddClient(id string, c *Client) error {
	_, err := s.kv.Get(id)
	if err != nil && err != defs.ErrNotFound {
		return err
	}

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
	if err != nil {
		if err != defs.ErrNotFound {
			return nil
		}
		return err
	}

	c := &Client{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return err
	}

	if c.Pending {
		return errors.New("client pending")
	}

	return s.kv.Del(id)
}

func (s *Storage) GetClient(id string) *Client {
	d, err := s.kv.Get(id)
	if err != nil {
		if err != defs.ErrNotFound {
			return nil
		}
		return nil
	}

	c := &Client{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return nil
	}

	if c.Pending {
		return nil
	}

	return nil
}
