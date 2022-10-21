package util

import (
	"go.uber.org/zap"

	"gitlab.qredo.com/computational-custodian/signing-agent/config"
)

func NewLogger(cfg *config.Logging) *zap.SugaredLogger {
	logConfig := zap.NewProductionConfig()

	if cfg.Format == "text" {
		logConfig = zap.NewDevelopmentConfig()
	}

	logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	if cfg.Level == "info" {
		logConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logConfig.DisableStacktrace = true
	l, _ := logConfig.Build()

	return l.Sugar()
}

func NewTestLogger() *zap.SugaredLogger {
	testConfig := &config.Logging{
		Format: "text",
		Level:  "info",
	}
	return NewLogger(testConfig)
}
