package util

import (
	"crypto/rand"
	"encoding/json"

	"github.com/pkg/errors"
	"gitlab.qredo.com/qredo-server/qredo-core/qerr"

	"github.com/qredo/assets/libs/crypto"
)

const (
	amclRandomSeedSize = 48
)

func randomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)

	return b, err
}

// CreateAMCLRng creates a new AMCL RNG with a random seed
func CreateAMCLRng() (*crypto.Rand, error) {
	b, err := randomBytes(amclRandomSeedSize)
	if err != nil {
		return nil, err
	}

	return crypto.NewRand(b), nil
}

func BLSSign(seed, payload []byte) ([]byte, error) {

	_, blsSecret, err := crypto.BLSKeys(crypto.NewRand(seed), nil)
	if err != nil {
		return nil, qerr.Wrap(err).WithMessage("generate BLS key")
	}

	signature, err := crypto.BLSSign(payload, blsSecret)
	if err != nil {
		return nil, errors.Wrap(err, "BLS sign payload")
	}

	return signature, nil
}

func BLSVerify(seed, msg, sig []byte) error {

	blsPublic, _, err := crypto.BLSKeys(crypto.NewRand(seed), nil)
	if err != nil {
		return errors.New("generate BLS key")
	}

	return crypto.BLSVerify(msg, blsPublic, sig)
}

func ZKPToken(zkpID, zkpToken []byte, pin int) ([]byte, error) {

	rng, err := CreateAMCLRng()
	if err != nil {
		return nil, qerr.Wrap(err)
	}
	zkpOnePass, err := crypto.ClientOnePass(zkpID, pin, rng, zkpToken, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Generate ZKP One Pass")
	}

	return json.Marshal(zkpOnePass)
}
