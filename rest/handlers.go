package rest

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"

	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/rest/version"
	"gitlab.qredo.com/custody-engine/automated-approver/util"

	"gitlab.qredo.com/custody-engine/automated-approver/defs"
	"gitlab.qredo.com/custody-engine/automated-approver/lib"
)

type handler struct {
	core      lib.SigningAgentClient
	cfg       config.Config
	log       *zap.SugaredLogger
	version   *version.Version
	websocket api.WebsocketStatus
	redis     *redis.Client
	rsync     *redsync.Redsync
}

func NewHandler(core lib.SigningAgentClient, config *config.Config, log *zap.SugaredLogger,
	version *version.Version, redis *redis.Client, rsync *redsync.Redsync) *handler {

	localFeedUrl := fmt.Sprintf("ws://%s%s/client/feed", config.HTTP.Addr, PathPrefix)
	remoteFeedUrl := genWSQredoCoreClientFeedURL(&config.Base)

	h := &handler{
		core:      core,
		cfg:       *config,
		log:       log,
		version:   version,
		websocket: api.NewWebsocketStatus(ConnectionState.Closed, remoteFeedUrl, localFeedUrl),
		redis:     redis,
		rsync:     rsync,
	}

	return h
}

// genWSQredoCoreClientFeedURL assembles and returns the Qredo WS client feed URL as a string.
func genWSQredoCoreClientFeedURL(config_base *config.Base) string {
	builder := strings.Builder{}
	builder.WriteString(config_base.WsScheme)
	builder.WriteString(config_base.QredoAPIDomain)
	builder.WriteString(config_base.QredoAPIBasePath)
	builder.WriteString("/coreclient/feed")
	return builder.String()
}

func (h *handler) UpdateWebsocketStatus(status string) {
	h.websocket.ReadyState = status
}

func (h *handler) GetWSQredoCoreClientFeedURL() string {
	return h.websocket.RemoteFeedUrl
}

// HealthCheckVersion
//
// swagger:route GET /healthcheck/version
//
// Check application version.
func (h *handler) HealthCheckVersion(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	response := h.version
	return response, nil
}

// HealthCheckConfig
//
// swagger:route GET /healthcheck/config
//
// Check application version.
func (h *handler) HealthCheckConfig(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	response := h.cfg
	return response, nil
}

// HealthCheckStatus
//
// swagger:route GET /healthcheck/status
//
// Check application status.
func (h *handler) HealthCheckStatus(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")

	response := api.HealthCheckStatusResponse{
		WebsocketStatus: h.websocket,
	}
	return response, nil
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

// AutoApprovalFunction
func (h *handler) AutoApproval() error {
	// enable auto-approval only if configured
	if !h.cfg.Base.AutoApprove {
		h.log.Debug("Auto-approval feature not enabled in config")
		return nil
	}

	h.log.Debug("Handler for Auto-approval background job")
	go AutoApproveHandler(h)

	return nil
}

// ClientFeed
//
// swagger:route POST /client/feed  ClientFeed
//
// Get approval requests Feed (via websocket) from Qredo Backend
func (h *handler) ClientFeed(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	h.log.Debug("Handler for ClientFeed endpoint")
	WebSocketFeedHandler(h, w, r)
	return nil, nil
}

// ClientFullRegister
//
// swagger:route POST /client/register ClientFullRegister
//
// Client registration process (3 steps in one)
//
// Responses:
//
//	200: ClientRegisterFinishResponse
func (h *handler) ClientFullRegister(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	h.log.Debug("Handler for ClientFullRegister endpoint")

	w.Header().Set("Content-Type", "application/json")

	if h.core.GetSystemAgentID() != "" {
		return nil, defs.ErrBadRequest().WithDetail("AgentID already exist. You can not set new one.")
	}
	response := api.ClientFullRegisterResponse{}
	cRegReq := &api.ClientRegisterRequest{}
	err := util.DecodeRequest(cRegReq, r)
	if err != nil {
		return nil, err
	}
	if err := cRegReq.Validate(); err != nil {
		return nil, defs.ErrBadRequest().WithDetail(err.Error())
	}
	registerResults, err := h.core.ClientRegister(cRegReq.Name) // we get bls, ec publicks keys
	if err != nil {
		return nil, err
	}

	reqDataInit := &api.QredoRegisterInitRequest{
		Name:         cRegReq.Name,
		BLSPublicKey: registerResults.BLSPublicKey,
		ECPublicKey:  registerResults.ECPublicKey,
	}

	initResults, err := h.core.ClientInit(reqDataInit, registerResults.RefID, cRegReq.APIKey, cRegReq.Base64PrivateKey)
	if err != nil {
		return response, err
	}

	response.AgentID = initResults.AccountCode
	reqDataFinish := &api.ClientRegisterFinishRequest{}
	copier.Copy(&reqDataFinish, &initResults) // initResults contain only one field more - timestamp
	_, err = h.core.ClientRegisterFinish(reqDataFinish, registerResults.RefID)
	if err != nil {
		return response, err
	}

	err = h.core.SetSystemAgentID(initResults.AccountCode)
	if err != nil {
		h.log.Errorf("Could not set AgentID to Storage: %s", err)
	}

	// return local feedUrl for request approvals
	response.FeedURL = h.websocket.LocalFeedUrl

	// also enable auto-approval of requests
	h.AutoApproval()

	return response, nil
}
