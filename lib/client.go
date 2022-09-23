package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/btcec"

	"github.com/pkg/errors"

	"github.com/google/uuid"

	"gitlab.qredo.com/custody-engine/automated-approver/util"

	"gitlab.qredo.com/custody-engine/automated-approver/api"

	"gitlab.qredo.com/custody-engine/automated-approver/crypto"
	defs "gitlab.qredo.com/custody-engine/automated-approver/defs"
)

func (h *signingAgent) ClientRegister(name string) (*api.ClientRegisterResponse, error) {

	var err error

	client := &Agent{Name: name}

	client.BLSSeed, err = util.RandomBytes(util.AMCLRandomSeedSize)
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

func (h *signingAgent) ClientRegisterFinish(req *api.ClientRegisterFinishRequest, ref string) (*api.ClientRegisterFinishResponse, error) {

	pending := h.store.GetPending(ref)
	if pending == nil {
		return nil, defs.ErrNotFound().WithDetail("ref_id").Wrap(errors.New("pending client not found"))
	}
	pending.ID = req.ID
	pending.AccountCode = req.AccountCode

	var err error
	pending.ZKPID, err = hex.DecodeString(req.ClientID) // this ClientID is a sensitive data
	if err != nil {
		return nil, errors.Wrap(err, "invalid sensitive data - ClientID in response")
	}
	cs, err := hex.DecodeString(req.ClientSecret)
	if err != nil {
		return nil, errors.Wrap(err, "invalid sensitive data - ClientSecret in response")
	}

	// ZKP Token
	pending.ZKPToken, err = crypto.ExtractPIN(pending.ZKPID, h.cfg.PIN, cs)
	if err != nil {
		return nil, errors.Wrap(err, "extract pin")
	}

	idDocRaw, err := hex.DecodeString(req.IDDocument)
	if err != nil {
		return nil, errors.Wrap(err, "invalid id document in response")
	}

	idDocSignature, err := util.BLSSign(pending.BLSSeed, idDocRaw)
	if err != nil {
		return nil, errors.Wrap(err, "idDoc sign")
	}

	zkpOnePass, err := util.ZKPOnePass(pending.ZKPID, pending.ZKPToken, h.cfg.PIN)
	if err != nil {
		return nil, errors.Wrap(err, "get zkp token")
	}

	confirmRequest := api.CoreClientServiceRegisterFinishRequest{
		IDDocSignatureHex: hex.EncodeToString(idDocSignature),
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpOnePass))

	finishResp := &api.CoreClientServiceRegisterFinishResponse{}

	if err = h.htc.Request(http.MethodPost, util.URLRegisterConfirm(h.cfg.HttpScheme, h.cfg.QredoAPIDomain, h.cfg.QredoAPIBasePath), confirmRequest, finishResp, header); err != nil {
		return nil, err
	}

	err = h.store.RemovePending(ref)
	if err != nil {
		return nil, err
	}

	err = h.store.AddAgent(pending.ID, pending)
	if err != nil {
		return nil, err
	}

	err = h.store.SetSystemAgentID(req.AccountCode)
	if err != nil {
		return nil, err
	}

	return &api.ClientRegisterFinishResponse{
		FeedURL: finishResp.Feed,
	}, nil
}

// ClientsList - Signing Agent can be only one
func (h *signingAgent) ClientsList() ([]string, error) {
	agentID := h.store.GetSystemAgentID()
	if len(agentID) > 0 {
		return []string{agentID}, nil
	} else {
		return []string{}, nil
	}
}

func (h *signingAgent) ClientInit(reqData *api.QredoRegisterInitRequest, ref, apikey, b64PrivateKey string) (*api.QredoRegisterInitResponse, error) {
	reqDataBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}
	req := &Request{Body: reqDataBody}
	GenTimestamp(req)
	err = DecodeBase64RSAKey(req, b64PrivateKey)
	if err != nil {
		return nil, err
	}
	req.ApiKey = strings.TrimSpace(apikey)
	req.Uri = util.URLClientInit(h.cfg.HttpScheme, h.cfg.QredoAPIDomain, h.cfg.QredoAPIBasePath)
	err = SignRequest(req)
	if err != nil {
		return nil, err
	}
	headers := GetClientInitHttpHeaders(req)

	var respData *api.QredoRegisterInitResponse = &api.QredoRegisterInitResponse{}
	if err = h.htc.Request(http.MethodPost, req.Uri, reqData, respData, headers); err != nil {
		return nil, err
	}
	return respData, nil
}

func (h *signingAgent) SetSystemAgentID(agetID string) error {
	return h.store.SetSystemAgentID(agetID)
}

func (h *signingAgent) GetSystemAgentID() string {
	return h.store.GetSystemAgentID()
}

func (h *signingAgent) GetAgentZKPOnePass() ([]byte, error) {
	agentID := h.store.GetSystemAgentID()
	if agentID == "" {
		return nil, errors.Errorf("can not get system agent ID from the store.")
	}
	agent := h.store.GetAgent(agentID)
	if agent == nil {
		return nil, errors.Errorf("can not get agent from the store.")
	}
	zkpOnePass, err := util.ZKPOnePass(agent.ZKPID, agent.ZKPToken, h.cfg.PIN)
	if err != nil {
		return nil, err
	}
	return zkpOnePass, nil
}
