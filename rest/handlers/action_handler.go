package handlers

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gitlab.qredo.com/computational-custodian/signing-agent/autoapprover"
	"gitlab.qredo.com/computational-custodian/signing-agent/defs"
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
// swagger:route PUT /client/action/{action_id}  actions actionApprove
//
// Approve action
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
// swagger:route DELETE /client/action/{action_id}  actions actionReject
//
// Reject action
func (h *ActionHandler) ActionReject(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	actionID = strings.TrimSpace(actionID)
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("empty actionID")
	}
	return nil, h.actionManager.Reject(actionID)
}
