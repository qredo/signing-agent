package lib

import (
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/core-client/util"
	"gitlab.qredo.com/qredo-server/qredo-core/qerr"

	"gitlab.qredo.com/qredo-server/core-client/defs"
)

// Sign signs a payload
func (h *coreClient) Sign(ctx *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := &api.SignRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	clientID := mux.Vars(r)["client_id"]
	if clientID == "" {
		return nil, qerr.BadRequest().WithReason("clientID")
	}

	msg, err := hex.DecodeString(req.MessageHashHex)
	if err != nil {
		return nil, qerr.BadRequest().WithReason("invalid_message_hex").Wrap(err).WithMessage("invalid message hex")
	}

	if len(msg) > 64 {
		return nil, qerr.BadRequest().WithReason("invalid_message_hex_size").Wrap(err).WithMessage("invalid message hex size %s ", strconv.Itoa(len(msg)))
	}

	client := h.store.GetClient(clientID)
	if client == nil {
		return nil, qerr.NotFound().WithReason("core_client_seed").Wrap(err).WithMessage("get lib client seed from secrets store %s", clientID)
	}

	signature, err := util.BLSSign(client.BLSSeed, msg)
	if err != nil {
		return nil, qerr.Internal().Wrap(err).WithMessage("sign message for lib client %s", clientID)
	}

	return &api.SignResponse{
		SignatureHex: hex.EncodeToString(signature),
		SignerID:     clientID,
	}, nil
}

// PartnerCompanySign -
func (h *coreClient) Verify(ctx *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := &api.VerifyRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	msg, err := hex.DecodeString(req.MessageHashHex)
	if err != nil {
		return nil, qerr.BadRequest().WithReason("invalid_message_hex").Wrap(err).WithMessage("invalid message hex")
	}
	if len(msg) > 64 {
		return nil, qerr.BadRequest().WithReason("invalid_message_hex_size").Wrap(err).WithMessage("invalid message hex size " + strconv.Itoa(len(msg)))
	}

	sig, err := hex.DecodeString(req.SignatureHex)
	if err != nil {
		return nil, qerr.BadRequest().WithReason("invalid_signature_hex").Wrap(err).WithMessage("invalid message hex")
	}

	client := h.store.GetClient(req.SignerID)
	if client == nil {
		return nil, qerr.NotFound().WithReason("signer_not_found").Wrap(err).WithMessage("get signer %s", req.SignerID)
	}

	if err := util.BLSVerify(client.BLSSeed, msg, sig); err != nil {
		return nil, qerr.Forbidden().WithReason("invalid_signature").Wrap(err).WithMessage("invalid signature")
	}

	return nil, nil
}
