package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/defs"
)

type AWSStore struct {
	lock       sync.RWMutex
	secretName string
	region     string
	svc        *secretsmanager.SecretsManager
}

// NewAWSStore creates and return the AWS KVStore.
func NewAWSStore(cfg config.AWSConfig) KVStore {
	s := &AWSStore{
		region:     cfg.Region,
		secretName: cfg.SecretName,
		lock:       sync.RWMutex{},
	}

	return s
}

// Get returns the value of the named key, or error if not found.
func (s *AWSStore) Get(key string) ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	secretData, err := s.getSecret(s.secretName)
	if err != nil {
		fmt.Println("error returned from getSecret:", err)
		return nil, err
	}

	var cfg map[string][]byte = make(map[string][]byte)
	if len(secretData) > 0 {
		err = json.Unmarshal(secretData, &cfg)
		if err != nil {
			return nil, err
		}
	}

	if val, ok := cfg[key]; ok {
		return val, nil
	}

	return nil, defs.KVErrNotFound
}

// Set adds/updates the named key with value in data.
func (s *AWSStore) Set(key string, data []byte) error {
	s.lock.RLock()

	secretData, err := s.getSecret(s.secretName)
	if err != nil {
		s.lock.RUnlock()
		return err
	}

	var cfg map[string][]byte = make(map[string][]byte)
	if len(secretData) > 0 {
		err = json.Unmarshal(secretData, &cfg)
		if err != nil {
			s.lock.RUnlock()
			return err
		}
	}

	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()

	cfg[key] = data

	secretData, err = json.Marshal(cfg)
	if err != nil {
		return err
	}

	s.setSecret(s.secretName, secretData)

	return nil
}

// Del deletes the named key.
func (s *AWSStore) Del(key string) error {
	s.lock.RLock()

	secretData, err := s.getSecret(s.secretName)
	if err != nil {
		s.lock.RUnlock()
		return err
	}

	var cfg map[string][]byte = make(map[string][]byte)
	if len(secretData) > 0 {
		err = json.Unmarshal(secretData, &cfg)
		if err != nil {
			s.lock.RUnlock()
			return err
		}
	}

	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(cfg, key)

	secretData, err = json.Marshal(cfg)
	if err != nil {
		return err
	}

	s.setSecret(s.secretName, secretData)

	return nil
}

// Init sets up the AWS session and checks the connection by reading the secret.
func (s *AWSStore) Init() error {
	// create an AWS SecretsManager client
	sess, err := session.NewSession()
	if err != nil {
		return err
	}
	s.svc = secretsmanager.New(sess, aws.NewConfig().WithRegion(s.region))

	// check connection
	_, err = s.getSecret(s.secretName)
	if err != nil {
		return errors.Wrap(err, "cannot initialise AWS store")
	}

	return nil
}

// getSecret reads the secret with name from AWS.  Various sanity checks on AWS access, returning errors.
// The secret should be binary and base64 encoded.  The decoded []byte is returned.
func (s *AWSStore) getSecret(name string) ([]byte, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := s.svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				return nil, errors.Wrap(err, "secret cannot be decypted using provided KMS key")
			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				return nil, errors.Wrap(err, "internal server error")
			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				return nil, errors.Wrap(err, "invalid parameter")
			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				return nil, errors.Wrap(err, "invalid parameter in current resource state")
			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				return nil, errors.Wrap(err, "resource not found")
			}
		}
		// not a known aws error
		return nil, err
	}

	// Decrypts secret using the associated KMS key.
	// We're expecting a binary. Check and return an error if a string is found.
	var decodedBinarySecretBytes []byte
	if result.SecretString != nil {
		return nil, fmt.Errorf("string returned, expected []byte")
	} else {
		decodedBinarySecretBytes = make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		_, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			return nil, err
		}
	}

	return decodedBinarySecretBytes, nil
}

// setSecret base64 encodes the data and stores it in the named secret.
func (s *AWSStore) setSecret(name string, data []byte) error {

	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)

	input := &secretsmanager.UpdateSecretInput{
		SecretBinary: dst,
		SecretId:     aws.String(name),
	}

	_, err := s.svc.UpdateSecret(input)
	if err != nil {
		return err
	}

	return nil
}
