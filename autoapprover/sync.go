package autoapprover

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
)

type cache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

type mockCache struct {
	GetCalled     bool
	SetCalled     bool
	LastKey       string
	NextStringCmd *redis.StringCmd
	NextStatusCmd *redis.StatusCmd
	LastValue     interface{}
}

func (m *mockCache) Get(ctx context.Context, key string) *redis.StringCmd {
	m.GetCalled = true
	m.LastKey = key
	return m.NextStringCmd
}
func (m *mockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	m.SetCalled = true
	m.LastKey = key
	m.LastValue = value
	return m.NextStatusCmd
}

type syncI interface {
	NewMutex(name string, options ...redsync.Option) *redsync.Mutex
}

type mockSync struct {
	NewMutexCalled bool
	LastName       string
	NextMutex      *redsync.Mutex
}

func (m *mockSync) NewMutex(name string, options ...redsync.Option) *redsync.Mutex {
	m.NewMutexCalled = true
	m.LastName = name
	return m.NextMutex
}

type mutex interface {
	Lock() error
	Unlock() (bool, error)
}

type mockMutex struct {
	NextLockError   error
	NextUnlockError error
	NextLock        bool
	LockCalled      bool
	UnlockCalled    bool
}

func (m *mockMutex) Lock() error {
	m.LockCalled = true
	return m.NextLockError
}

func (m *mockMutex) Unlock() (bool, error) {
	m.UnlockCalled = true

	return m.NextLock, m.NextUnlockError
}
