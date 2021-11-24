package handlers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

func NewStore(name string) (*Storage, error) {
	s := &Storage{
		fileName: name,
		Pending:  map[string]*Client{},
		Clients:  map[string]*Client{},
	}
	if err := s.Load(); err != nil {
		return nil, err
	}

	return s, nil
}

type Client struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	BLSSeed     []byte `json:"bls_seed"`
	AccountCode string `json:"account_code,omitempty"`
	ZKPID       []byte `json:"zkpid,omitempty"`
	ZKPToken    []byte `json:"zkptoken,omitempty"`
}

type Storage struct {
	fileName string
	mtx      sync.RWMutex
	Pending  map[string]*Client `json:"pending"`
	Clients  map[string]*Client `json:"clients"`
}

func (s *Storage) Load() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	b, err := ioutil.ReadFile(s.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return s.save()
		}
		return err
	}

	if err := json.Unmarshal(b, s); err != nil {
		return err
	}

	return nil
}

// caller must handle concurrency
func (s *Storage) save() error {

	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.fileName, b, 0600)
}

func (s *Storage) AddPending(ref string, c *Client) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.Pending[ref] = c
	return s.save()
}

func (s *Storage) RemovePending(ref string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.Pending, ref)

	return s.save()
}

func (s *Storage) GetPending(ref string) *Client {
	if c, ok := s.Pending[ref]; ok {
		return c
	}

	return nil
}

func (s *Storage) AddClient(id string, c *Client) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.Clients[id] = c

	return s.save()
}

func (s *Storage) RemoveClient(id string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.Clients, id)

	return s.save()
}

func (s *Storage) GetClient(id string) *Client {
	if c, ok := s.Clients[id]; ok {
		return c
	}

	return nil
}
