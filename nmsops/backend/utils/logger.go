package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {

	logConfig := zap.NewDevelopmentConfig()

	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	Logger = zap.Must(logConfig.Build()) // New development for current basis

}
