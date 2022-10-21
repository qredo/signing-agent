package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"gitlab.qredo.com/computational-custodian/signing-agent/config"
	"gitlab.qredo.com/computational-custodian/signing-agent/defs"
	"gitlab.qredo.com/computational-custodian/signing-agent/lib"
)

var rMutex *redsync.Mutex
var rCtx = context.Background()

type handler struct {
	core  lib.SigningAgentClient
	cfg   config.Config
	log   *zap.SugaredLogger
	redis *redis.Client
	rsync *redsync.Redsync
}

func NewHandler(core lib.SigningAgentClient, config *config.Config, log *zap.SugaredLogger, redis *redis.Client, rsync *redsync.Redsync) *handler {

	h := &handler{
		core:  core,
		cfg:   *config,
		log:   log,
		redis: redis,
		rsync: rsync,
	}

	return h
}

// ClientsList
//
// swagger:route GET /client ClientsList
//
// # Return AgentID if it's configured
//
// Responses:
//
//	200: []string
func (h *handler) ClientsList(_ *defs.RequestContext, w http.ResponseWriter, _ *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	return h.core.ClientsList()
}

// ActionApprove
//
// swagger:route PUT /client/action/{action_id}  actions actionApprove
//
// Approve action
func (h *handler) ActionApprove(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("actionID")
	}

	if h.cfg.LoadBalancing.Enable {
		if err := h.redis.Get(rCtx, actionID).Err(); err == nil {
			h.log.Debugf("action [%v] was already approved!", actionID)
			return nil, nil
		}
		rMutex = h.rsync.NewMutex(actionID)
		if err := rMutex.Lock(); err != nil {
			time.Sleep(time.Duration(h.cfg.LoadBalancing.OnLockErrorTimeOutMs) * time.Millisecond)
			return nil, err
		}
		defer func() {
			if ok, err := rMutex.Unlock(); !ok || err != nil {
				h.log.Errorf("%v action-id %v", err, actionID)
			}
			h.redis.Set(rCtx, actionID, 1, time.Duration(h.cfg.LoadBalancing.ActionIDExpirationSec)*time.Second)
		}()
	}

	return nil, h.core.ActionApprove(actionID)
}

// ActionReject
//
// swagger:route DELETE /client/action/{action_id}  actions actionReject
//
// Reject action
func (h *handler) ActionReject(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("actionID")
	}
	return nil, h.core.ActionReject(actionID)
}
