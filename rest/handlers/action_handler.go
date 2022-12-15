package handlers

import (
	"net/http"
	"strings"

	"github.com/qredo/signing-agent/autoapprover"
	"github.com/qredo/signing-agent/defs"

	"github.com/gorilla/mux"
)

type ActionHandler struct {
	actionManager autoapprover.ActionManager
}

func NewActionHandler(actionManager autoapprover.ActionManager) *ActionHandler {
	return &ActionHandler{
		actionManager: actionManager,
	}
}

// ActionApprove
//
// swagger:route PUT /client/action/{action_id} action ActionApprove
//
// # Approve a transaction
//
// This endpoint approves a transaction based on the transaction ID, `action_id`, passed.
//
//	Parameters:
//	  + name: action_id
//	    in: path
//	    description: the ID of the transaction that is received from the feed
//	    required: true
//	    type: string
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GenericResponse
func (h *ActionHandler) ActionApprove(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	actionID = strings.TrimSpace(actionID)
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("empty actionID")
	}

	return nil, h.actionManager.Approve(actionID)
}

// ActionReject
//
// swagger:route DELETE /client/action/{action_id} action actionReject
//
// # Reject a transaction
//
// This endpoint rejects a transaction based on the transaction ID, `action_id`, passed.
//
//	Parameters:
//	  + name: action_id
//	    in: path
//	    description: the ID of the transaction that is received from the feed
//	    required: true
//	    type: string
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GenericResponse
func (h *ActionHandler) ActionReject(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	actionID = strings.TrimSpace(actionID)
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("empty actionID")
	}
	return nil, h.actionManager.Reject(actionID)
}
