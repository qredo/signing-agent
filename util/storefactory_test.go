package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
)

func Test_StoreFactory_CreateStore_Creates_file_store(t *testing.T) {
	// Arrange
	cfg := config.Base{
		StoreType: "file",
	}

	// Act
	sut := CreateStore(cfg)

	// Assert
	require.NotNil(t, sut)
	assert.Equal(t, reflect.TypeOf(sut).String(), "*util.FileStore")
}

func Test_StoreFactory_CreateStore_Creates_oci_store(t *testing.T) {
	// Arrange
	cfg := config.Base{
		StoreType: "oci",
	}

	// Act
	sut := CreateStore(cfg)

	// Assert
	require.NotNil(t, sut)
	assert.Equal(t, reflect.TypeOf(sut).String(), "*util.OciStore")
}

func Test_StoreFactory_CreateStore_Creates_aws_store(t *testing.T) {
	// Arrange
	cfg := config.Base{
		StoreType: "aws",
	}

	// Act
	sut := CreateStore(cfg)

	// Assert
	require.NotNil(t, sut)
	assert.Equal(t, reflect.TypeOf(sut).String(), "*util.AWSStore")
}
