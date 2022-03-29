package lib

import (
	"encoding/hex"
	"net/http"

	"github.com/pkg/errors"

	"gitlab.qredo.com/qredo-server/qredo-core/api/partner"
	"gitlab.qredo.com/qredo-server/qredo-core/qerr"

	"gitlab.qredo.com/qredo-server/core-client/util"

	"gitlab.qredo.com/qredo-server/core-client/defs"
)

func (h *coreClient) ActionApprove(clientID, actionID string) error {
	client := h.store.GetClient(clientID)
	if client == nil {
		return qerr.NotFound().WithReason("client_id")
	}

	zkpToken, err := util.ZKPToken(client.ZKPID, client.ZKPToken, h.cfg.PIN)
	if err != nil {
		return errors.Wrap(err, "get zkp token")
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))
	messagesResp := &partner.CoreClientServiceActionMessagesResponse{}
	if err = h.htc.Request(http.MethodGet, util.URLActionMessages(h.cfg.QredoServerURL, actionID), nil, messagesResp, header); err != nil {
		return qerr.Wrap(err)
	}

	if messagesResp.Messages == nil {
		return qerr.NotFound().WithDetails("messages")
	}

	signatures := make([]string, len(messagesResp.Messages))

	for i, m := range messagesResp.Messages {
		msg, err := hex.DecodeString(m)
		if err != nil || len(msg) == 0 {
			return qerr.Wrap(err)
		}

		signature, err := util.BLSSign(client.BLSSeed, msg)
		if err != nil {
			return qerr.Wrap(err)
		}
		signatures[i] = hex.EncodeToString(signature)
	}

	zkpToken, err = util.ZKPToken(client.ZKPID, client.ZKPToken, h.cfg.PIN)
	if err != nil {
		return errors.Wrap(err, "get zkp token")
	}

	req := &partner.CoreClientServiceActionApproveRequest{
		Signatures: signatures,
	}
	header = http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))
	if err = h.htc.Request(http.MethodPut, util.URLActionApprove(h.cfg.QredoServerURL, actionID), req, nil, header); err != nil {
		return qerr.Wrap(err)
	}

	return nil
}

func (h *coreClient) ActionReject(clientID, actionID string) error {
	client := h.store.GetClient(clientID)
	if client == nil {
		return qerr.NotFound().WithReason("client_id")
	}

	zkpToken, err := util.ZKPToken(client.ZKPID, client.ZKPToken, h.cfg.PIN)
	if err != nil {
		return errors.Wrap(err, "get zkp token")
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))

	if err = h.htc.Request(http.MethodDelete, util.URLActionReject(h.cfg.QredoServerURL, actionID), nil, nil, header); err != nil {
		return qerr.Wrap(err)
	}

	return nil
}
