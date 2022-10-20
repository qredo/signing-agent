// Package autoapprover provides a mechanism to receive action information as bytes.
// The action data is analyzed and if it meets the requirements, the action is approved.
// It supports approval retrying based on defined intervals

package autoapprover

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/hub"
	"gitlab.qredo.com/custody-engine/automated-approver/lib"

	"go.uber.org/zap"
)

var defaultCtx = context.Background()

type AutoApprover struct {
	hub.FeedClient
	log              *zap.SugaredLogger
	cfgLoadBalancing *config.LoadBalancing
	cfgAutoApproval  *config.AutoApprove
	cache            cache
	mutex            mutex
	sync             syncI
	core             lib.SigningAgentClient
	lastError        error
}

// NewAutoApprover returns a new *AutoApprover instance initialized with the provided parameters
// The AutoApprover has an internal FeedClient which means it will be stopped when the service stops
// or the Feed channel is closed on the sender side
func NewAutoApprover(core lib.SigningAgentClient, log *zap.SugaredLogger, config *config.Config, cache cache, sync syncI) *AutoApprover {
	return &AutoApprover{
		FeedClient:       hub.NewFeedClient(true),
		log:              log,
		cfgLoadBalancing: &config.LoadBalancing,
		cfgAutoApproval:  &config.AutoApprove,
		cache:            cache,
		sync:             sync,
		core:             core,
	}
}

// Listen is constantly listening for messages on the Feed channel.
// The Feed channel is always closed by the sender. Then this happens, the AutoApprover stops
func (a *AutoApprover) Listen() {
	for {
		if message, ok := <-a.Feed; !ok {
			//channel was closed by the sender
			a.log.Info("AutoApproval: stopped")
			return
		} else {
			a.handleMessage(message)
		}
	}
}

func (a *AutoApprover) handleMessage(message []byte) {
	var action actionInfo
	if err := json.Unmarshal(message, &action); err == nil {
		if action.IsNotExpired() {
			if a.shouldHandleAction(action.ID) {
				go a.handleAction(&action)
			}
		} else {
			a.log.Infof("AutoApproval: action [%v] has expired", action.ID)
		}
	} else {
		a.log.Errorf("AutoApproval: error [%v] while unmarshaling the message [%v]", err, string(message))
		a.lastError = err
	}
}

func (a *AutoApprover) shouldHandleAction(actionId string) bool {
	if a.cfgLoadBalancing.Enable {
		//check if the action was already picked up by another signing agent
		if err := a.cache.Get(defaultCtx, actionId).Err(); err == nil {
			a.log.Debugf("AutoApproval: action [%v] was already approved!", actionId)
			return false
		}
		a.mutex = a.sync.NewMutex(actionId)
	}
	return true
}

func (a *AutoApprover) handleAction(action *actionInfo) error {
	if a.cfgLoadBalancing.Enable {
		if err := a.mutex.Lock(); err != nil {
			time.Sleep(time.Duration(a.cfgLoadBalancing.OnLockErrorTimeOutMs) * time.Millisecond)
			return err
		}
		defer func() {
			if ok, err := a.mutex.Unlock(); !ok || err != nil {
				a.log.Errorf("AutoApproval: %v action [%v]", err, action.ID)
			}
			a.cache.Set(defaultCtx, action.ID, 1, time.Duration(a.cfgLoadBalancing.ActionIDExpirationSec)*time.Second)
		}()
	}

	return a.approveAction(action.ID, action.AgentID)
}

func (a *AutoApprover) approveAction(actionId, agentId string) error {
	timer := newRetryTimer(a.cfgAutoApproval.RetryInterval, a.cfgAutoApproval.RetryIntervalMax)
	for {
		if err := a.core.ActionApprove(actionId); err == nil {
			a.log.Infof("AutoApproval: action [%v] approved automatically", actionId)
			return nil
		} else {
			a.log.Errorf("AutoApproval: approval failed for [actionID:%v]. Error msg: %v", agentId, actionId, err)

			if timer.isTimeOut() {
				a.log.Warnf("AutoApproval: auto action approve failed [actionID:%v]", actionId)
				return errors.New("timeout")
			}

			a.log.Warnf("AutoApproval: auto approve action is repeated [actionID:%v] ", actionId)
			timer.retry()
		}
	}
}
