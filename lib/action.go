package lib

import (
	"encoding/hex"
	"net/http"

	"gitlab.qredo.com/custody-engine/automated-approver/api"

	"github.com/pkg/errors"

	"gitlab.qredo.com/custody-engine/automated-approver/util"

	"gitlab.qredo.com/custody-engine/automated-approver/defs"
)

func (h *autoApprover) ActionApprove(agentID, actionID string) error {
	agent := h.store.GetClient(agentID)
	if agent == nil {
		return defs.ErrNotFound().WithDetail("agentID")
	}

	zkpToken, err := util.ZKPToken(agent.ZKPID, agent.ZKPToken, h.cfg.PIN)
	if err != nil {
		return errors.Wrap(err, "get zkp token")
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))
	messagesResp := &api.CoreClientServiceActionMessagesResponse{}
	if err = h.htc.Request(http.MethodGet, util.URLActionMessages(h.cfg.QredoURL, actionID), nil, messagesResp, header); err != nil {
		return err
	}

	if messagesResp.Messages == nil || len(messagesResp.Messages) == 0 {
		return defs.ErrNotFound().WithDetail("messages")
	}

	signatures := make([]string, len(messagesResp.Messages))

	for i, m := range messagesResp.Messages {
		msg, err := hex.DecodeString(m)
		if err != nil || len(msg) == 0 {
			return err
		}

		signature, err := util.BLSSign(agent.BLSSeed, msg)
		if err != nil {
			return err
		}
		signatures[i] = hex.EncodeToString(signature)
	}

	zkpToken, err = util.ZKPToken(agent.ZKPID, agent.ZKPToken, h.cfg.PIN)
	if err != nil {
		return errors.Wrap(err, "get zkp token")
	}

	req := &api.CoreClientServiceActionApproveRequest{
		Signatures: signatures,
	}
	header = http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))
	if err = h.htc.Request(http.MethodPut, util.URLActionApprove(h.cfg.QredoURL, actionID), req, nil, header); err != nil {
		return err
	}

	return nil
}

func (h *autoApprover) ActionReject(agentID, actionID string) error {
	agent := h.store.GetClient(agentID)
	if agent == nil {
		return defs.ErrNotFound().WithDetail("agentID")
	}

	zkpToken, err := util.ZKPToken(agent.ZKPID, agent.ZKPToken, h.cfg.PIN)
	if err != nil {
		return errors.Wrap(err, "get zkp token")
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))

	if err = h.htc.Request(http.MethodDelete, util.URLActionReject(h.cfg.QredoURL, actionID), nil, nil, header); err != nil {
		return err
	}

	return nil
}
