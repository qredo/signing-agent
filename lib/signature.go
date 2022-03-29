package lib

import (
	"encoding/hex"
	"strconv"

	"gitlab.qredo.com/qredo-server/core-client/api"
	"gitlab.qredo.com/qredo-server/core-client/util"
	"gitlab.qredo.com/qredo-server/qredo-core/qerr"
)

// Sign signs a payload
func (h *coreClient) Sign(clientID, messageHex string) (*api.SignResponse, error) {
	msg, err := hex.DecodeString(messageHex)
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

// Verify verifies a signature
func (h *coreClient) Verify(req *api.VerifyRequest) error {
	msg, err := hex.DecodeString(req.MessageHashHex)
	if err != nil {
		return qerr.BadRequest().WithReason("invalid_message_hex").Wrap(err).WithMessage("invalid message hex")
	}
	if len(msg) > 64 {
		return qerr.BadRequest().WithReason("invalid_message_hex_size").Wrap(err).WithMessage("invalid message hex size " + strconv.Itoa(len(msg)))
	}

	sig, err := hex.DecodeString(req.SignatureHex)
	if err != nil {
		return qerr.BadRequest().WithReason("invalid_signature_hex").Wrap(err).WithMessage("invalid message hex")
	}

	client := h.store.GetClient(req.SignerID)
	if client == nil {
		return qerr.NotFound().WithReason("signer_not_found").Wrap(err).WithMessage("get signer %s", req.SignerID)
	}

	if err := util.BLSVerify(client.BLSSeed, msg, sig); err != nil {
		return qerr.Forbidden().WithReason("invalid_signature").Wrap(err).WithMessage("invalid signature")
	}

	return nil
}
