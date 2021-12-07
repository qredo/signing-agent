package handlers

import (
	"encoding/hex"
	"net/http"

	"github.com/pkg/errors"
	"gitlab.qredo.com/qredo-server/qredo-core/qerr"

	"gitlab.qredo.com/qredo-server/qredo-core/qapi/partner"

	"github.com/gorilla/mux"
	"gitlab.qredo.com/qredo-server/core-client/util"

	"gitlab.qredo.com/qredo-server/core-client/defs"
)

func (h *Handler) ActionApprove(ctx *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, qerr.BadRequest().WithReason("actionID")
	}
	clientID := mux.Vars(r)["client_id"]
	if clientID == "" {
		return nil, qerr.BadRequest().WithReason("clientID")
	}

	client := h.store.GetClient(clientID)
	if client == nil {
		return nil, qerr.NotFound().WithReason("client_id")
	}

	zkpToken, err := util.ZKPToken(client.ZKPID, client.ZKPToken, h.cfg.PIN)
	if err != nil {
		return nil, errors.Wrap(err, "get zkp token")
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))
	messagesResp := &partner.ServiceActionMessagesResponse{}
	if err = h.htc.Request(http.MethodGet, util.URLActionMessages(h.cfg.QredoServerURL, actionID), nil, messagesResp, header); err != nil {
		return nil, qerr.Wrap(err)
	}

	if messagesResp.Messages == nil {
		return nil, qerr.NotFound().WithDetails("messages")
	}

	signatures := make([]string, len(messagesResp.Messages))

	for i, m := range messagesResp.Messages {
		msg, err := hex.DecodeString(m)
		if err != nil || len(msg) == 0 {
			return nil, qerr.Wrap(err)
		}

		signature, err := util.BLSSign(msg, client.BLSSeed)
		if err != nil {
			return nil, qerr.Wrap(err)
		}
		signatures[i] = hex.EncodeToString(signature)
	}

	zkpToken, err = util.ZKPToken(client.ZKPID, client.ZKPToken, h.cfg.PIN)
	if err != nil {
		return nil, errors.Wrap(err, "get zkp token")
	}

	req := &partner.ServiceActionApproveRequest{
		Signatures: signatures,
	}
	header = http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))
	if err = h.htc.Request(http.MethodGet, util.URLActionApprove(h.cfg.QredoServerURL, actionID), req, nil, header); err != nil {
		return nil, qerr.Wrap(err)
	}

	return nil, nil
}

func (h *Handler) ActionReject(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	actionID := mux.Vars(r)["action_id"]
	if actionID == "" {
		return nil, qerr.BadRequest().WithReason("actionID")
	}
	clientID := mux.Vars(r)["client_id"]
	if clientID == "" {
		return nil, qerr.BadRequest().WithReason("clientID")
	}

	client := h.store.GetClient(clientID)
	if client == nil {
		return nil, qerr.NotFound().WithReason("client_id")
	}

	zkpToken, err := util.ZKPToken(client.ZKPID, client.ZKPToken, h.cfg.PIN)
	if err != nil {
		return nil, errors.Wrap(err, "get zkp token")
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))

	if err = h.htc.Request(http.MethodGet, util.URLActionReject(h.cfg.QredoServerURL, actionID), nil, nil, header); err != nil {
		return nil, qerr.Wrap(err)
	}

	return nil, nil
}
