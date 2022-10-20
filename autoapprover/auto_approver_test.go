package autoapprover

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/test-go/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/hub"
	"gitlab.qredo.com/custody-engine/automated-approver/lib"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
	"go.uber.org/goleak"
)

func TestAutoApprover_Listen_fails_to_unmarshal(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mock_core := &lib.MockSigningAgentClient{}
	sut := &AutoApprover{
		core:       mock_core,
		FeedClient: hub.NewFeedClient(true),
		log:        util.NewTestLogger(),
	}
	defer close(sut.Feed)
	go sut.Listen()

	//Act
	sut.Feed <- []byte("")
	<-time.After(time.Second) //give it time to finish

	//Assert
	assert.NotNil(t, sut.lastError)
	assert.Equal(t, "unexpected end of JSON input", sut.lastError.Error())
	assert.False(t, mock_core.ActionApproveCalled)
}

func TestAutoApprover_handleMessage_action_expired(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	sut := &AutoApprover{
		log: util.NewTestLogger(),
	}
	bytes, _ := json.Marshal(actionInfo{
		ExpireTime: 12360,
	})

	//Act
	sut.handleMessage(bytes)

	//Assert
	assert.Nil(t, sut.lastError)
}

func TestAutoApprover_shouldHandleAction_cached(t *testing.T) {
	//Arrange
	cacheMock := &mockCache{
		NextStringCmd: redis.NewStringCmd(context.Background()),
	}
	sut := NewAutoApprover(nil, util.NewTestLogger(), &config.Config{LoadBalancing: config.LoadBalancing{Enable: true}}, cacheMock, nil)
	bytes, _ := json.Marshal(actionInfo{
		ID:         "actionid",
		ExpireTime: time.Now().Add(time.Minute).Unix(),
	})

	//Act
	sut.handleMessage(bytes)

	//Assert
	assert.True(t, cacheMock.GetCalled)
	assert.Equal(t, "actionid", cacheMock.LastKey)
}

func TestAutoApprover_shouldHandleAction_gets_mutex(t *testing.T) {
	//Arrange
	syncMock := &mockSync{
		NextMutex: &redsync.Mutex{},
	}
	stringCmd := redis.NewStringCmd(context.Background())
	stringCmd.SetErr(errors.New("some error"))
	cacheMock := &mockCache{
		NextStringCmd: stringCmd,
	}

	sut := &AutoApprover{
		cfgLoadBalancing: &config.LoadBalancing{
			Enable: true,
		},
		sync:  syncMock,
		cache: cacheMock,
	}

	//Act
	res := sut.shouldHandleAction("actionid")

	//Assert
	assert.True(t, res)
	assert.True(t, cacheMock.GetCalled)
	assert.Equal(t, "actionid", cacheMock.LastKey)
	assert.True(t, syncMock.NewMutexCalled)
	assert.Equal(t, "actionid", syncMock.LastName)
	assert.NotNil(t, sut.mutex)
}

func TestAutoApprover_handleAction_lock_error(t *testing.T) {
	//Arrange
	mutexMock := &mockMutex{
		NextLockError: errors.New("some lock error"),
	}

	sut := &AutoApprover{
		cfgLoadBalancing: &config.LoadBalancing{
			Enable: true,
		},
		mutex: mutexMock,
	}

	//Act
	err := sut.handleAction(nil)

	//Assert
	assert.NotNil(t, err)
	assert.Contains(t, "some lock error", err.Error())
	assert.True(t, mutexMock.LockCalled)
}

func TestAutoApprover_handleAction_unlock_error(t *testing.T) {
	//Arrange
	stringCmd := redis.NewStringCmd(context.Background())
	stringCmd.SetErr(errors.New("some error"))
	cacheMock := &mockCache{
		NextStringCmd: stringCmd,
	}
	mutexMock := &mockMutex{
		NextLock:        true,
		NextUnlockError: errors.New("some unlock error"),
	}

	sut := &AutoApprover{
		log: util.NewTestLogger(),
		cfgLoadBalancing: &config.LoadBalancing{
			Enable: true,
		},
		cfgAutoApproval: &config.AutoApprove{},
		cache:           cacheMock,
		mutex:           mutexMock,
		core:            &lib.MockSigningAgentClient{},
	}

	//Act
	err := sut.handleAction(&actionInfo{
		ID:         "some action id",
		ExpireTime: time.Now().Add(time.Minute).Unix(),
	})

	//Assert
	assert.Nil(t, err)
	assert.True(t, cacheMock.SetCalled)
	assert.Equal(t, "some action id", cacheMock.LastKey)
	assert.Equal(t, 1, cacheMock.LastValue)
	assert.True(t, mutexMock.UnlockCalled)
}

func TestAutoApprover_approveAction_times_out(t *testing.T) {
	//Arrange
	coreMock := &lib.MockSigningAgentClient{
		NextError: errors.New("some error"),
	}
	sut := &AutoApprover{
		core: coreMock,
		cfgAutoApproval: &config.AutoApprove{
			RetryIntervalMax: 2,
			RetryInterval:    1,
		},
		log: util.NewTestLogger(),
	}

	//Act
	err := sut.approveAction("some action id", "some agent id")

	//Assert
	assert.NotNil(t, err)
	assert.Equal(t, "timeout", err.Error())
	assert.True(t, coreMock.ActionApproveCalled)
	assert.Equal(t, "some action id", coreMock.LastActionId)
}
