package logging

import (
	"go.uber.org/zap"
)

type environment byte

const (
	EnvDevelopment environment = iota
	EnvProduction
)

func NewWithEnv(env environment) *zap.Logger {
	var logger *zap.Logger

	switch env {
	case EnvDevelopment:
		logger, _ = zap.NewDevelopment()
	case EnvProduction:
		logger, _ = zap.NewProduction()
	}

	return logger
}
