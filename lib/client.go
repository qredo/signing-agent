package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/btcsuite/btcd/btcec"

	"github.com/pkg/errors"

	"github.com/google/uuid"

	"gitlab.qredo.com/qredo-server/core-client/util"

	"gitlab.qredo.com/qredo-server/core-client/api"

	"gitlab.qredo.com/qredo-server/core-client/crypto"
	"gitlab.qredo.com/qredo-server/core-client/defs"
)

func (h *coreClient) ClientRegister(name string) (*api.ClientRegisterResponse, error) {

	var err error

	client := &Client{Name: name}

	client.BLSSeed, err = util.RandomBytes(48)
	if err != nil {
		return nil, err
	}

	// EC Public key
	hashedSeed := sha256.Sum256(client.BLSSeed)
	_, cpub1 := btcec.PrivKeyFromBytes(btcec.S256(), hashedSeed[:])
	ecPublicKey := cpub1.SerializeUncompressed()

	refID := uuid.New().String()

	if err = h.store.AddPending(refID, client); err != nil {
		return nil, err
	}

	blsPublic, _, err := crypto.BLSKeys(crypto.NewRand(client.BLSSeed), nil)
	if err != nil {
		return nil, errors.Wrap(err, "generate BLS key")
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
		return nil, defs.ErrNotFound().WithDetail("ref_id").Wrap(errors.New("pending client not found"))
	}
	pending.ID = req.ID
	pending.AccountCode = req.AccountCode

	var err error
	pending.ZKPID, err = hex.DecodeString(req.ClientID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid client id in response")
	}
	cs, err := hex.DecodeString(req.ClientSecret)
	if err != nil {
		return nil, errors.Wrap(err, "invalid client id in response")
	}

	// ZKP Token
	pending.ZKPToken, err = crypto.ExtractPIN(pending.ZKPID, h.cfg.PIN, cs)
	if err != nil {
		return nil, errors.Wrap(err, "extract pin")
	}

	idDocRaw, err := hex.DecodeString(req.IDDoc)
	if err != nil {
		return nil, errors.Wrap(err, "invalid id document in response")
	}

	idDocSignature, err := util.BLSSign(pending.BLSSeed, idDocRaw)
	if err != nil {
		return nil, errors.Wrap(err, "idDoc sign")
	}

	zkpToken, err := util.ZKPToken(pending.ZKPID, pending.ZKPToken, h.cfg.PIN)
	if err != nil {
		return nil, errors.Wrap(err, "get zkp token")
	}

	confirmRequest := api.CoreClientServiceRegisterFinishRequest{
		IDDocSignatureHex: hex.EncodeToString(idDocSignature),
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))

	finishResp := &api.CoreClientServiceRegisterFinishResponse{}

	if err = h.htc.Request(http.MethodPost, util.URLRegisterConfirm(h.cfg.QredoURL), confirmRequest, finishResp, header); err != nil {
		return nil, err
	}

	err = h.store.RemovePending(ref)
	if err != nil {
		return nil, err
	}

	err = h.store.AddClient(pending.ID, pending)
	if err != nil {
		return nil, err
	}

	return &api.ClientRegisterFinishResponse{
		FeedURL: finishResp.Feed,
	}, nil
}
func (h *coreClient) ClientsList() (interface{}, error) {
	return "not implemented", nil
}
