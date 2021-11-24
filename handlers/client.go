package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/btcsuite/btcd/btcec"

	"gitlab.qredo.com/qredo-server/qredo-core/qapi/coreclient"

	"github.com/pkg/errors"

	"github.com/gorilla/mux"

	"github.com/google/uuid"

	"gitlab.qredo.com/qredo-server/core-client/util"

	"gitlab.qredo.com/qredo-server/core-client/api"

	"gitlab.qredo.com/qredo-server/qredo-core/qerr"

	"github.com/qredo/assets/libs/crypto"
	"gitlab.qredo.com/qredo-server/qredo-core/qcommon"

	"gitlab.qredo.com/qredo-server/core-client/defs"
)

func (h *Handler) ClientRegister(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := &api.ClientRegisterRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	client := &Client{Name: req.Name}

	client.BLSSeed, err = qcommon.RandomBytes(48)
	if err != nil {
		return nil, qerr.Wrap(err)
	}

	// EC Public key
	hashedSeed := sha256.Sum256(client.BLSSeed)
	_, cpub1 := btcec.PrivKeyFromBytes(btcec.S256(), hashedSeed[:])
	ecPublicKey := cpub1.SerializeUncompressed()

	refID := uuid.New().String()

	if err = h.store.AddPending(refID, client); err != nil {
		return nil, qerr.Wrap(err)
	}

	blsPublic, _, err := crypto.BLSKeys(crypto.NewRand(client.BLSSeed), nil)
	if err != nil {
		return nil, qerr.Wrap(err).WithMessage("generate BLS key")
	}

	return &api.ClientRegisterResponse{
		BLSPublicKey: hex.EncodeToString(blsPublic),
		ECPublicKey:  hex.EncodeToString(ecPublicKey),
		RefID:        refID,
	}, nil

}

func (h *Handler) ClientRegisterFinish(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	ref := mux.Vars(r)["ref"]
	req := &api.ClientRegisterFinishRequest{}
	err := util.DecodeRequest(req, r)
	if err != nil {
		return nil, err
	}

	pending := h.store.GetPending(ref)
	if pending == nil {
		return nil, qerr.NotFound().Wrap(err).WithMessage("pending client not found").WithDetails("ref_id")
	}
	pending.ID = req.ID
	pending.AccountCode = req.AccountCode

	pending.ZKPID, err = hex.DecodeString(req.ClientID)
	if err != nil {
		return nil, qerr.Wrap(err).WithMessage("invalid client id in response")
	}
	cs, err := hex.DecodeString(req.ClientSecret)
	if err != nil {
		return nil, qerr.Wrap(err).WithMessage("invalid client id in response")
	}

	// ZKP Token
	pending.ZKPToken, err = crypto.ExtractPIN(pending.ZKPID, h.cfg.PIN, cs)
	if err != nil {
		return nil, qerr.Wrap(err).WithMessage("extract pin")
	}

	idDocRaw, err := hex.DecodeString(req.IDDoc)
	if err != nil {
		return nil, qerr.Wrap(err).WithMessage("invalid id document in response")
	}

	idDocSignature, err := util.BLSSign(pending.BLSSeed, idDocRaw)
	if err != nil {
		return nil, qerr.Wrap(err).WithMessage("idDoc sign")
	}

	zkpToken, err := util.ZKPToken(pending.ZKPID, pending.ZKPToken, h.cfg.PIN)
	if err != nil {
		return nil, errors.Wrap(err, "get zkp token")
	}

	confirmRequest := coreclient.RegisterFinishRequest{
		IDDocSignatureHex: hex.EncodeToString(idDocSignature),
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))

	finishResp := &coreclient.RegisterFinishResponse{}

	if err = h.htc.Request(http.MethodPost, util.URLRegisterConfirm(h.cfg.QredoServerURL, pending.ID), confirmRequest, finishResp, header); err != nil {
		return nil, qerr.Wrap(err)
	}

	err = h.store.RemovePending(ref)
	if err != nil {
		return nil, qerr.Wrap(err)
	}

	err = h.store.AddClient(pending.ID, pending)
	if err != nil {
		return nil, qerr.Wrap(err)
	}

	return &api.ClientRegisterFinishResponse{
		FeedURL: finishResp.Feed,
	}, nil
}
func (h *Handler) ClientsList(s *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (interface{}, error) {
	return nil, nil
}
