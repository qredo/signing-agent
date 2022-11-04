package handlers

import (
	"net/http"
	"sync"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"

	"signing-agent/api"
	"signing-agent/autoapprover"
	"signing-agent/clientfeed"
	"signing-agent/config"
	"signing-agent/defs"
	"signing-agent/hub"
	"signing-agent/lib"
	"signing-agent/util"
)

type newClientFeedFunc func(conn hub.WebsocketConnection, log *zap.SugaredLogger, unregister clientfeed.UnregisterFunc, config *config.WebSocketConf) clientfeed.ClientFeed

type SigningAgentHandler struct {
	feedHub           hub.FeedHub
	log               *zap.SugaredLogger
	core              lib.SigningAgentClient
	config            *config.AutoApprove
	websocketConfig   *config.WebSocketConf
	localFeed         string
	decode            func(interface{}, *http.Request) error
	autoApprover      *autoapprover.AutoApprover
	upgrader          hub.WebsocketUpgrader
	newClientFeedFunc newClientFeedFunc //function used by the feed clients to unregister themselves from the hub and stop receiving data
}

// NewSigningAgentHandler instantiates and returns a new SigningAgentHandler object.
func NewSigningAgentHandler(feedHub hub.FeedHub, core lib.SigningAgentClient, log *zap.SugaredLogger, config *config.Config, autoApprover *autoapprover.AutoApprover, upgrader hub.WebsocketUpgrader, localFeed string) *SigningAgentHandler {
	return &SigningAgentHandler{
		feedHub:           feedHub,
		log:               log,
		core:              core,
		config:            &config.AutoApprove,
		localFeed:         localFeed,
		decode:            util.DecodeRequest,
		autoApprover:      autoApprover,
		upgrader:          upgrader,
		websocketConfig:   &config.Websocket,
		newClientFeedFunc: clientfeed.NewClientFeed,
	}
}

// StartAgent is running the feed hub if the agent is registered.
// It also makes sure the auto approver is registered to the hub and is listening for incoming actions, if enabled in the config
func (h *SigningAgentHandler) StartAgent() {
	agentID := h.core.GetSystemAgentID()
	if len(agentID) == 0 {
		h.log.Info("Agent is not yet configured, auto-approval not started")
		return
	}

	if !h.feedHub.Run() {
		h.log.Error("failed to start the feed hub")
		return
	}

	if !h.config.Enabled {
		h.log.Debug("Auto-approval feature not enabled in config")
		return
	}

	h.feedHub.RegisterClient(&h.autoApprover.FeedClient)
	go h.autoApprover.Listen()
}

// StopAgent is called to stop the feed hub on request, by ex: when the service is stopped
func (h *SigningAgentHandler) StopAgent() {
	h.feedHub.Stop()
	h.log.Info("feed hub stopped")
}

// RegisterAgent
//
// swagger:route POST /client/register RegisterAgent
//
// Client registration process (3 steps in one)
//
// Responses:
//
// 200: ClientFullRegisterResponse
func (h *SigningAgentHandler) RegisterAgent(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")

	if h.core.GetSystemAgentID() != "" {
		return nil, defs.ErrBadRequest().WithDetail("AgentID already exist. You can not set new one.")
	}

	if response, err := h.register(r); err != nil {
		return nil, err
	} else {
		h.StartAgent()
		return response, nil
	}
}

// ClientFeed
//
// swagger:route POST /client/feed  ClientFeed
//
// Get approval requests Feed (via websocket) from Qredo Backend
func (h *SigningAgentHandler) ClientFeed(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	hubRunning := h.feedHub.IsRunning()

	if hubRunning {
		clientFeed := h.newClientFeed(w, r)
		if clientFeed != nil {
			var wg sync.WaitGroup
			wg.Add(1)

			go clientFeed.Start(&wg)
			wg.Wait() //wait for the client to set up the conn handling

			h.feedHub.RegisterClient(clientFeed.GetFeedClient())
			go clientFeed.Listen()
			h.log.Info("handler: connected to feed, listening ...")
		}
	} else {
		h.log.Debugf("handler: failed to connect, hub not running")
	}

	return nil, nil
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
func (h *SigningAgentHandler) ClientsList(_ *defs.RequestContext, w http.ResponseWriter, _ *http.Request) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	return h.core.ClientsList()
}

func (h *SigningAgentHandler) newClientFeed(w http.ResponseWriter, r *http.Request) clientfeed.ClientFeed {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Errorf("handler: failed to upgrade connection, err: %v", err)
		return nil
	}

	return h.newClientFeedFunc(conn, h.log, h.feedHub.UnregisterClient, h.websocketConfig)
}

func (h *SigningAgentHandler) register(r *http.Request) (interface{}, error) {
	registerRequest, err := h.validateRegisterRequest(r)
	if err != nil {
		return nil, err
	}

	registerResults, err := h.core.ClientRegister(registerRequest.Name) // Get BLS and EC public keys
	if err != nil {
		h.log.Debugf("error while trying to register the client [%s], err: %v", registerRequest.Name, err)
		return nil, err
	}

	initResults, err := h.initRegistration(registerResults, registerRequest)
	if err != nil {
		h.log.Debugf("error while trying to init the client registration, err: %v", err)
		return nil, err
	}

	if err := h.finishRegistration(initResults, registerResults.RefID); err != nil {
		return nil, err
	}

	response := api.ClientFullRegisterResponse{
		AgentID: initResults.AccountCode,
		FeedURL: h.localFeed,
	}

	return response, nil
}

func (h *SigningAgentHandler) validateRegisterRequest(r *http.Request) (*api.ClientRegisterRequest, error) {
	register := &api.ClientRegisterRequest{}
	if err := h.decode(register, r); err != nil {
		h.log.Debugf("failed to decode register request, %v", err)
		return nil, err
	}

	if err := register.Validate(); err != nil {
		h.log.Debugf("failed to validate register request, %v", err)
		return nil, defs.ErrBadRequest().WithDetail(err.Error())
	}

	return register, nil
}

func (h *SigningAgentHandler) initRegistration(register *api.ClientRegisterResponse, reqData *api.ClientRegisterRequest) (*api.QredoRegisterInitResponse, error) {
	reqDataInit := api.NewQredoRegisterInitRequest(reqData.Name, register.BLSPublicKey, register.ECPublicKey)
	return h.core.ClientInit(reqDataInit, register.RefID, reqData.APIKey, reqData.Base64PrivateKey)
}

func (h *SigningAgentHandler) finishRegistration(initResults *api.QredoRegisterInitResponse, refId string) error {
	reqDataFinish := &api.ClientRegisterFinishRequest{}

	// initResults contains only one extra field, timestamp
	if err := copier.Copy(&reqDataFinish, &initResults); err != nil {
		return err
	}

	if _, err := h.core.ClientRegisterFinish(reqDataFinish, refId); err != nil {
		h.log.Debugf("error while finishing client registration, %v", err)
		return err
	}

	return nil
}
