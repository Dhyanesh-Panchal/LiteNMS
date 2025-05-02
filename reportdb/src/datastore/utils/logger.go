package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
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
			"./logs/prod_" + time.Now().Format("2006_01_02") + ".log",
		}

		Logger = zap.Must(prodConfig.Build())

	} else {

		devConfig := zap.NewDevelopmentConfig()

		devConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		devConfig.Level.SetLevel(zapcore.DebugLevel)

		devConfig.OutputPaths = []string{
			"stdout",
			"./logs/dev_" + time.Now().Format("2006_01_02") + ".log",
		}

		Logger = zap.Must(devConfig.Build())

	}

	return nil

}
