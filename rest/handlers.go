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

// ClientsList
//
// swagger:route GET /client clientsList ClientsList
//
// Return AgentID if it's configured
//
// Responses:
//      200: []string
func (h *handler) ClientsList(_ *defs.RequestContext, _ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	return h.core.ClientsList()
}

// ActionApprove
//
// swagger:route PUT /client/action/{action_id}  actions actionApprove
//
// Approve action
//
func (h *handler) ActionApprove(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("actionID")
	}

	return nil, h.core.ActionApprove(actionID)
}

// ActionReject
//
// swagger:route DELETE /client/action/{action_id}  actions actionReject
//
// Reject action
//
func (h *handler) ActionReject(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("actionID")
	}
	return nil, h.core.ActionReject(actionID)
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

	agentID := mux.Vars(r)["agent_id"]
	if agentID == "" {
		return nil, defs.ErrBadRequest().WithDetail("agentID")
	}

	return h.core.Sign(agentID, req.MessageHashHex)
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
	go WebSocketHandler(h)

	return nil
}

// ClientFeed
//
// swagger:route POST /client/feed  clientFeed ClientFeed
//
// Get approval requests Feed (via websocket) from Qredo Backend
//
func (h *handler) ClientFeed(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	h.log.Debug("Handler for ClientFeed endpoint")
	WebSocketFeedHandler(h, w, r)
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
	response.FeedURL = fmt.Sprintf("ws://%s%s/client/feed", h.cfg.HTTP.Addr, pathPrefix)

	// also enable auto-approval of requests
	h.AutoApproval()

	return response, nil
}
