package autoapprover

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/lib"
	"gitlab.qredo.com/custody-engine/automated-approver/websocket"
	"go.uber.org/zap"
)

var defaultCtx = context.Background()

type AutoApproval struct {
	websocket.FeedClient
	log           *zap.SugaredLogger
	loadBalancing *config.LoadBalancing
	autoApproval  *config.AutoApprove
	cache         cache
	mutex         mutex
	sync          syncI
	core          lib.SigningAgentClient
	lastError     error
}

func NewAutoApproval(core lib.SigningAgentClient, log *zap.SugaredLogger, config *config.Config, cache cache, sync syncI) *AutoApproval {
	return &AutoApproval{
		FeedClient:    websocket.NewFeedClient(true),
		log:           log,
		loadBalancing: &config.LoadBalancing,
		autoApproval:  &config.AutoApprove,
		cache:         cache,
		sync:          sync,
		core:          core,
	}
}

func (a *AutoApproval) Listen() {
	for {
		if message, ok := <-a.Feed; !ok {
			//channel was closed by the feed hub due to disconnected websocket to qredo server
			a.log.Info("auto approval stopped")
			return
		} else {
			a.handleMessage(message)
		}
	}
}

func (a *AutoApproval) handleMessage(message []byte) {
	var action ActionInfo
	if err := json.Unmarshal(message, &action); err == nil {
		if action.IsNotExpired() {
			if a.shouldHandleAction(action.ID) {
				go a.handleAction(&action)
			}
		} else {
			a.log.Infof("action [%v] has expired", action.ID)
		}
	} else {
		a.log.Errorf("error [%v] while unmarshaling the message [%v]", err, string(message))
		a.lastError = err
	}
}

func (a *AutoApproval) shouldHandleAction(actionId string) bool {
	if a.loadBalancing.Enable {
		if err := a.cache.Get(defaultCtx, actionId).Err(); err == nil {
			a.log.Debugf("action [%v] was already approved!", actionId)
			return false
		}
		a.mutex = a.sync.NewMutex(actionId)
	}
	return true
}

func (a *AutoApproval) handleAction(action *ActionInfo) error {
	if a.loadBalancing.Enable {
		if err := a.mutex.Lock(); err != nil {
			time.Sleep(time.Duration(a.loadBalancing.OnLockErrorTimeOutMs) * time.Millisecond)
			return err
		}
		defer func() {
			if ok, err := a.mutex.Unlock(); !ok || err != nil {
				a.log.Errorf("%v action [%v]", err, action.ID)
			}
			a.cache.Set(defaultCtx, action.ID, 1, time.Duration(a.loadBalancing.ActionIDExpirationSec)*time.Second)
		}()
	}

	return a.approveAction(action.ID, action.AgentID)
}

func (a *AutoApproval) approveAction(actionId, agentId string) error {
	timer := NewRetryTimer(a.autoApproval.RetryInterval, a.autoApproval.RetryIntervalMax)
	for {
		if err := a.core.ActionApprove(actionId); err == nil {
			a.log.Infof("Action [%v] approved automatically", actionId)
			return nil
		} else {
			a.log.Errorf("Approval failed for [actionID:%v]. Error msg: %v", agentId, actionId, err)

			if timer.isTimeOut() {
				a.log.Warnf("Auto action approve failed [actionID:%v]", actionId)
				return errors.New("timeout")
			}

			a.log.Warnf("Auto approve action is repeated [actionID:%v] ", actionId)
			timer.retry()
		}
	}
}
