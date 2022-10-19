package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/jinzhu/copier"
	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/autoapprover"
	"gitlab.qredo.com/custody-engine/automated-approver/clientfeed"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/defs"
	"gitlab.qredo.com/custody-engine/automated-approver/hub"
	"gitlab.qredo.com/custody-engine/automated-approver/lib"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
	"go.uber.org/zap"
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
func NewSigningAgentHandler(feedHub hub.FeedHub, core lib.SigningAgentClient, log *zap.SugaredLogger, config *config.Config, autoApprover *autoapprover.AutoApprover, upgrader hub.WebsocketUpgrader) *SigningAgentHandler {
	return &SigningAgentHandler{
		feedHub:           feedHub,
		log:               log,
		core:              core,
		config:            &config.AutoApprove,
		localFeed:         fmt.Sprintf("ws://%s%s/client/feed", config.HTTP.Addr, defs.PathPrefix),
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

// RegisterAgent handles the registration and starting of a new agent
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

// ClientFeed provides a way to open a websocket connection to receive data send through the websocket connection with the Qredo back end
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
	copier.Copy(&reqDataFinish, &initResults) // initResults contains only one extra field, timestamp

	if _, err := h.core.ClientRegisterFinish(reqDataFinish, refId); err != nil {
		h.log.Debugf("error while finishing client registration, %v", err)
		return err
	}

	return nil
}
