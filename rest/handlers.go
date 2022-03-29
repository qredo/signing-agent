package rest

import (
	"net/http"

	"gitlab.qredo.com/qredo-server/qredo-core/qerr"

	"github.com/gorilla/mux"

	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/core-client/util"

	"gitlab.qredo.com/qredo-server/core-client/defs"
	"gitlab.qredo.com/qredo-server/core-client/lib"
)

type handler struct {
	core lib.CoreClient
}

// ClientRegister
//
// swagger:route POST /client browser clientRegister clientRegisterInit
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
// swagger:route POST /client{ref}  browser clientRegister clientRegisterFinish
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

// ClientsList
//
// swagger:route GET /client browser clientsList
//
// List registered clients
//
// Responses:
//      200: clientListResponse
func (h *handler) ClientsList(_ *defs.RequestContext, _ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	return h.core.ClientsList()
}

// ActionApprove
//
// swagger:route PUT /client/{client_id}/action/{action_id} browser actions actionApprove
//
// Approve action
//
// Responses:
//      200
func (h *handler) ActionApprove(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, qerr.BadRequest().WithReason("actionID")
	}
	clientID := mux.Vars(r)["client_id"]
	if clientID == "" {
		return nil, qerr.BadRequest().WithReason("clientID")
	}

	return nil, h.core.ActionApprove(clientID, actionID)
}

// ActionReject
//
// swagger:route DELETE /client/{client_id}/action/{action_id} browser actions actionReject
//
// Reject action
//
// Responses:
//      200
func (h *handler) ActionReject(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, qerr.BadRequest().WithReason("actionID")
	}
	clientID := mux.Vars(r)["client_id"]
	if clientID == "" {
		return nil, qerr.BadRequest().WithReason("clientID")
	}

	return nil, h.core.ActionReject(clientID, actionID)
}

// Sign
//
// swagger:route POST /client/{client_id}/sign browser payloads payloadSign
//
// Sign a payload
//
// Responses:
//      200: SignResponse
func (h *handler) Sign(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := &api.SignRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	clientID := mux.Vars(r)["client_id"]
	if clientID == "" {
		return nil, qerr.BadRequest().WithReason("clientID")
	}

	return h.core.Sign(clientID, req.MessageHashHex)
}

// Verify
//
// swagger:route POST /verify browser payloads signatureVerify
//
// Verify a signature
//
// Responses:
//      200
func (h *handler) Verify(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := &api.VerifyRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	return nil, h.core.Verify(req)
}
