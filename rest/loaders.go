package rest

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func loadRSAKey(req *request) error {
	f, err := os.Open(*flagPrivatePEMFilePath)
	if err != nil {
		return errors.Wrap(err, "load RSA key")
	}
	defer func() {
		err = f.Close()
		if err != nil {
			fmt.Println("cannot close open RSA key file: ", err)
		}
	}()

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
	k, err := os.Open(*flagAPIKeyFilePath)
	if err != nil {
		return errors.Wrap(err, "cannot open api key file")

	}
	defer func() {
		err = k.Close()
		if err != nil {
			fmt.Println("cannot close open apikey file: ", err)
		}
	}()

	key, err := ioutil.ReadAll(k)
	if err != nil {
		return errors.Wrap(err, "cannot read api key file")
	}

	req.apiKey = strings.TrimSpace(string(key))
	return nil
}
