package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Logger *zap.Logger

func InitLogger() error {

	err := os.MkdirAll("./logs/", os.ModePerm)

	if err != nil {

		return err

	}

	if IsProductionEnvironment {

		prodConfig := zap.NewProductionConfig()

		prodConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		prodConfig.Level.SetLevel(zapcore.ErrorLevel)

		prodConfig.OutputPaths = []string{
			"./logs/production.log",
		}

		Logger = zap.Must(prodConfig.Build())

	} else {

		devConfig := zap.NewDevelopmentConfig()

		devConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		devConfig.Level.SetLevel(zapcore.DebugLevel)

		devConfig.OutputPaths = []string{
			"stdout",
			"./logs/development.log",
		}

		Logger = zap.Must(devConfig.Build())

	}

	return nil

}
