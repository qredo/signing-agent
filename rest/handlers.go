package rest

import (
	"fmt"
	"net/http"

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
	core lib.AutomatedApproverClient
	cfg  config.Config
	log  *zap.SugaredLogger
}

//basic healthcheck response
type HealthcheckAlive struct {
	APIVersion    string `json:"api_version"`
	SchemaVersion string `json:"schema_version"`
}

// HealthCheck
//
// swagger:route GET /healthcheck
//
// Check if application is working.
//
func (h *handler) HealthCheck(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	response := HealthcheckAlive{
		APIVersion:    version.APIVersion,
		SchemaVersion: version.SchemaVersion,
	}
	return response, nil
}

// ClientRegister
//
// swagger:route POST /client clientRegister clientRegisterInit
//
// Initiate client registration procedure
//
// Responses:
//      200: clientRegisterResponse
func (h *handler) ClientRegister(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := &api.ClientRegisterRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	return h.core.ClientRegister(req.Name)
}

// ClientRegisterFinish
//
// swagger:route POST /client/{ref}  clientRegister clientRegisterFinish
//
// Finish client registration procedure
//
// Responses:
//      200: clientRegisterFinishResponse
func (h *handler) ClientRegisterFinish(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	ref := mux.Vars(r)["ref"]
	req := &api.ClientRegisterFinishRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	return h.core.ClientRegisterFinish(req, ref)
}

func (h *handler) ClientsList(_ *defs.RequestContext, _ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	return h.core.ClientsList()
}

// ActionApprove
//
// swagger:route PUT /client/{client_id}/action/{action_id}  actions actionApprove
//
// Approve action
//
func (h *handler) ActionApprove(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("actionID")
	}
	clientID := mux.Vars(r)["client_id"]
	if clientID == "" {
		return nil, defs.ErrBadRequest().WithDetail("clientID")
	}

	return nil, h.core.ActionApprove(clientID, actionID)
}

// ActionReject
//
// swagger:route DELETE /client/{client_id}/action/{action_id}  actions actionReject
//
// Reject action
//
func (h *handler) ActionReject(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("actionID")
	}
	clientID := mux.Vars(r)["client_id"]
	if clientID == "" {
		return nil, defs.ErrBadRequest().WithDetail("clientID")
	}

	return nil, h.core.ActionReject(clientID, actionID)
}

// Sign
//
// swagger:route POST /client/{client_id}/sign  payloads payloadSign
//
// Sign a payload
//
// Responses:
//      200: signResponse
func (h *handler) Sign(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := &api.SignRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	clientID := mux.Vars(r)["client_id"]
	if clientID == "" {
		return nil, defs.ErrBadRequest().WithDetail("clientID")
	}

	return h.core.Sign(clientID, req.MessageHashHex)
}

// Verify
//
// swagger:route POST /verify  payloads signatureVerify
//
// Verify a signature
//
func (h *handler) Verify(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := &api.VerifyRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	return nil, h.core.Verify(req)
}

// AutoApprovalFunction
//
func (h *handler) AutoApproval() error {
	// enable auto-approval only if configured
	if !h.cfg.Base.AutoApprove {
		h.log.Debug("Autoapproval feature not enabled in config")
		return nil
	}

	h.log.Debug("Handler for AutoApproval background job")

	var clientID string

	clientID = h.core.GetAgentID()
	if clientID == "" {
		h.log.Info("Agent is not yet configured, skipping Websocket connection for auto-approval")
		return nil
	}

	req := &lib.Request{}
	GenWSQredoCoreClientFeedURL(h, clientID, req)
	lib.GenTimestamp(req)

	err := lib.LoadRSAKey(req, h.cfg.Base.PrivatePEMFilePath)
	if err != nil {
		return err
	}
	h.log.Debugf("Loaded RSA key for AutoApproval from %s", h.cfg.Base.PrivatePEMFilePath)

	err = lib.LoadAPIKey(req, h.cfg.Base.APIKeyFilePath)
	if err != nil {
		return err
	}
	h.log.Debugf("Loaded API key for AutoApproval from %s", h.cfg.Base.APIKeyFilePath)

	err = lib.SignRequest(req)
	if err != nil {
		return err
	}

	go WebSocketHandler(h, req)

	return nil
}

// ClientFeed
//
// Get approval requests Feed (via websocket) from Qredo Backend
//
func (h *handler) ClientFeed(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	h.log.Debug("Handler for ClientFeed endpoint")
	coreClientID := mux.Vars(r)["client_id"] // also called AccoundID or AgentID
	if coreClientID == "" {
		return nil, defs.ErrBadRequest().WithDetail("coreClientID")
	}
	req := &lib.Request{}

	GenWSQredoCoreClientFeedURL(h, coreClientID, req)
	lib.GenTimestamp(req)
	err := lib.LoadRSAKey(req, h.cfg.Base.PrivatePEMFilePath)
	if err != nil {
		return nil, err
	}
	err = lib.LoadAPIKey(req, h.cfg.Base.APIKeyFilePath)
	if err != nil {
		return nil, err
	}
	err = lib.SignRequest(req)
	if err != nil {
		return nil, err
	}
	WebSocketFeedHandler(h, req, w, r)
	return nil, nil
}

// ClientFullRegister
//
// swagger:route POST /client/register  clientFullRegister ClientFullRegister
//
// Client registration process (3 steps in one)
//
// Responses:
//      200: ClientRegisterFinishResponse
func (h *handler) ClientFullRegister(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	h.log.Debug("Handler for ClientFullRegister endpoint")
	if h.core.GetAgentID() != "" {
		return nil, defs.ErrBadRequest().WithDetail("AgentID already exist. You can not set new one.")
	}
	response := api.ClientFullRegisterResponse{}
	req := &api.ClientRegisterRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}
	registerResults, err := h.core.ClientRegister(req.Name) // we get bls, ec publicks keys
	if err != nil {
		return nil, err
	}

	reqDataInit := &api.QredoRegisterInitRequest{
		Name:         req.Name,
		BLSPublicKey: registerResults.BLSPublicKey,
		ECPublicKey:  registerResults.ECPublicKey,
	}
	initResults, err := h.core.ClientInit(reqDataInit, registerResults.RefID)
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

	err = h.core.SetAgentID(initResults.AccountCode)
	if err != nil {
		h.log.Errorf("Could not set AgentID to Storage: %s", err)
	}

	// return local feedUrl for request approvals
	response.FeedURL = fmt.Sprintf("ws://%s%s/client/%s/feed", h.cfg.HTTP.Addr, pathPrefix, initResults.AccountCode)

	// also enable auto-approval of requests
	h.AutoApproval()

	return response, nil
}
