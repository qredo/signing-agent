package lib

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoaderAPIKey(t *testing.T) {

	t.Run(
		"Load correct APIKey",
		func(t *testing.T) {
			data := []byte("testapikeydata")
			err := os.WriteFile(TestDataAPIKeyFilePath, data, 0644)
			assert.NoError(t, err)
			defer func() {
				err = os.Remove(TestDataAPIKeyFilePath)
				assert.NoError(t, err)
			}()
			var req = &Request{}
			assert.Empty(t, req.ApiKey)
			LoadAPIKey(req, TestDataAPIKeyFilePath)
			assert.NotEmpty(t, req.ApiKey)
		})
	t.Run(
		"Load correct APIKey - file not found",
		func(t *testing.T) {
			var req = &Request{}
			assert.Empty(t, req.ApiKey)
			err := LoadAPIKey(req, TestDataAPIKeyFilePath)
			assert.Error(t, err)
			assert.Empty(t, req.ApiKey)
		})
}

func generatePrivateKey(t *testing.T, filePath string) {
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err, "Cannot generate RSA key")

	// dump generated private key to file
	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(privatekey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privatePem, err := os.Create(filePath)
	assert.NoError(t, err, "Cannot create private key file")

	err = pem.Encode(privatePem, privateKeyBlock)
	assert.NoError(t, err, "Cannot encode private key")
}

func TestLoaderRSAKey(t *testing.T) {
	generatePrivateKey(t, TestDataPrivatePEMFilePath)
	defer func() {
		os.Remove(TestDataPrivatePEMFilePath)
	}()
	t.Run(
		"Load correct RSAKey",
		func(t *testing.T) {
			var req = &Request{}
			assert.Empty(t, req.RsaKey)
			LoadRSAKey(req, TestDataPrivatePEMFilePath)
			assert.NotEmpty(t, req.RsaKey)
		})
	t.Run(
		"Load RSAKey - file not found",
		func(t *testing.T) {
			var req = &Request{}
			assert.Nil(t, req.RsaKey)
			LoadRSAKey(req, TestDataPrivatePEMFilePath+"faile_path")
			assert.Nil(t, req.RsaKey)
		})
}
