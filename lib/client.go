package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"gitlab.qredo.com/qredo-server/qredo-core/api/partner"
	"gitlab.qredo.com/qredo-server/qredo-core/qerr"

	"github.com/btcsuite/btcd/btcec"

	"github.com/pkg/errors"

	"github.com/google/uuid"

	"gitlab.qredo.com/qredo-server/core-client/util"

	"gitlab.qredo.com/qredo-server/core-client/api"

	"github.com/qredo/assets/libs/crypto"
	"gitlab.qredo.com/qredo-server/core-client/defs"
)

func (h *coreClient) ClientRegister(name string) (*api.ClientRegisterResponse, error) {

	var err error

	client := &Client{Name: name}

	client.BLSSeed, err = util.RandomBytes(48)
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

func (h *coreClient) ClientRegisterFinish(req *api.ClientRegisterFinishRequest, ref string) (*api.ClientRegisterFinishResponse, error) {

	pending := h.store.GetPending(ref)
	if pending == nil {
		return nil, qerr.NotFound().WithMessage("pending client not found").WithDetails("ref_id")
	}
	pending.ID = req.ID
	pending.AccountCode = req.AccountCode

	var err error
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

	confirmRequest := partner.CoreClientServiceRegisterFinishRequest{
		IDDocSignatureHex: hex.EncodeToString(idDocSignature),
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))

	finishResp := &partner.CoreClientServiceRegisterFinishResponse{}

	if err = h.htc.Request(http.MethodPost, util.URLRegisterConfirm(h.cfg.QredoServerURL), confirmRequest, finishResp, header); err != nil {
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
func (h *coreClient) ClientsList() (interface{}, error) {
	return "not implemented", nil
}
