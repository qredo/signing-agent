package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/core-client/rest/version"
	"gitlab.qredo.com/qredo-server/core-client/util"

	"gitlab.qredo.com/qredo-server/core-client/defs"
	"gitlab.qredo.com/qredo-server/core-client/lib"
)

type handler struct {
	core lib.CoreClient
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

// ClientFeed
//
// Get Core Client Feed (via websocket) from Qredo Backend
//
func (h *handler) ClientFeed(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	fmt.Printf("Handler for ClientFeed endpoint")
	coreClientID := mux.Vars(r)["client_id"] // also called AccoundID
	if coreClientID == "" {
		return nil, defs.ErrBadRequest().WithDetail("coreClientID")
	}
	req := &lib.Request{}

	genWSQredoCoreClientFeedURL(coreClientID, req)
	lib.GenTimestamp(req)
	err := lib.LoadRSAKey(req, *flagPrivatePEMFilePath)
	if err != nil {
		return nil, err
	}
	err = lib.LoadAPIKey(req, *flagAPIKeyFilePath)
	if err != nil {
		return nil, err
	}
	err = lib.SignRequest(req)
	if err != nil {
		return nil, err
	}
	webSocketHandler(h, req, w, r)
	return nil, nil
}

// ClientFullRegister
//
// swagger:route POST /client/register  clientFullRegister ClientFullRegister
//
// Finish client registration procedure (3 steps in one)
//
// Responses:
//      200: ClientRegisterFinishResponse
func (h *handler) ClientFullRegister(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	fmt.Printf("Handler for ClientFullRegister endpoint")
	response := api.ClientFullRegisterResponse{}
	req := &api.ClientRegisterRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}
	first_results, err := h.core.ClientRegister(req.Name) // we get bls, ec publicks keys
	if err != nil {
		return nil, err
	}

	response.ClientRegisterResponse = *first_results
	reqData := &api.QredoRegisterInitRequest{
		Name:         req.Name,
		BLSPublicKey: first_results.BLSPublicKey,
		ECPublicKey:  first_results.ECPublicKey,
	}
	second_results, err := h.core.ClientInit(reqData, first_results.RefID)
	if err != nil {
		return response, err
	}

	response.QredoRegisterInitResponse = *second_results
	reqData2 := &api.ClientRegisterFinishRequest{
		ID:           second_results.ID,
		AccountCode:  second_results.AccountCode,
		ClientID:     second_results.ClientID,
		ClientSecret: second_results.ClientSecret,
		IDDoc:        second_results.IDDocument,
	}
	finish_response, err := h.core.ClientRegisterFinish(reqData2, first_results.RefID)
	if err != nil {
		return response, err
	}
	response.ClientRegisterFinishResponse = *finish_response

	return response, nil
}
