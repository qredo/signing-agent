package rest

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func loadRSAKey(req *request) error {
	f, err := os.Open("/Users/leszek.kolacz/VSCodeProjects/core-client/private.pem")
	defer f.Close()

	if err != nil {
		return errors.Wrap(err, "load RSA key")
	}

	pemData, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "can't read RSA key file")
	}

	block, _ := pem.Decode(pemData)

	req.rsaKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return errors.Wrap(err, "parse RSA key")
	}
	return nil
}

func loadAPIKey(req *request) error {
	k, err := os.Open("/Users/leszek.kolacz/VSCodeProjects/core-client/apikey")
	defer k.Close()
	if err != nil {
		return errors.Wrap(err, "cannot open api key file")

	}

	key, err := ioutil.ReadAll(k)
	if err != nil {
		return errors.Wrap(err, "cannot read api key file")
	}

	req.apiKey = strings.TrimSpace(string(key))
	return nil
}
