package util

import (
	"gitlab.qredo.com/custody-engine/automated-approver/config"
)

func CreateStore(cfg config.Base) KVStore {
	switch cfg.StoreType {
	case "file":
		return NewFileStore(cfg.StoreFile)
	case "oci":
		return NewOciStore(cfg.StoreOci)
	case "aws":
		return NewAWSStore(cfg.StoreAWS)
	default:
		return nil
	}
}
