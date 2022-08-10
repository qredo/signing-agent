package lib

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func LoadRSAKey(req *Request, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "load RSA key")
	}
	defer f.Close()

	pemData, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "can't read RSA key file")
	}

	block, _ := pem.Decode(pemData)

	req.RsaKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return errors.Wrap(err, "parse RSA key")
	}
	return nil
}

func LoadAPIKey(req *Request, path string) error {
	k, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "cannot open api key file")

	}
	defer k.Close()

	key, err := ioutil.ReadAll(k)
	if err != nil {
		return errors.Wrap(err, "cannot read api key file")
	}

	req.ApiKey = strings.TrimSpace(string(key))
	return nil
}
