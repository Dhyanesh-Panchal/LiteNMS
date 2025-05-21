package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

var Logger *zap.Logger

func InitLogger() error {

	if err := os.MkdirAll("./logs/", os.ModePerm); err != nil {

		return err

	}

	if IsProductionEnvironment {

		encoderConfig := zap.NewProductionEncoderConfig()

		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		rotatingLogger := &lumberjack.Logger{

			Filename: "./logs/prod_" + time.Now().Format("2006_01_02") + ".log",

			MaxSize: MaxLogFileSizeInMB, // In MB

			MaxBackups: 3,

			MaxAge: LogFileRetentionInDays, // In days

			Compress: true,
		}

		// log level to error for production
		levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {

			return lvl >= zapcore.ErrorLevel

		})

		core := zapcore.NewCore(

			zapcore.NewJSONEncoder(encoderConfig),

			zapcore.AddSync(rotatingLogger),

			levelEnabler,
		)

		Logger = zap.New(core, zap.AddCaller())

	} else {

		encoderConfig := zap.NewDevelopmentEncoderConfig()

		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		// log level to debug for development
		levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {

			return lvl >= zapcore.DebugLevel

		})

		rotatingLogger := &lumberjack.Logger{

			Filename: "./logs/dev_" + time.Now().Format("2006_01_02") + ".log",

			MaxSize: MaxLogFileSizeInMB,

			MaxBackups: 3,

			MaxAge: LogFileRetentionInDays,

			Compress: false,
		}

		// Core for logging in console
		consoleCore := zapcore.NewCore(

			zapcore.NewConsoleEncoder(encoderConfig),

			zapcore.AddSync(os.Stdout),

			levelEnabler,
		)

		// Core for logging in file
		fileCore := zapcore.NewCore(

			zapcore.NewJSONEncoder(encoderConfig),

			zapcore.AddSync(rotatingLogger),

			levelEnabler,
		)

		core := zapcore.NewTee(consoleCore, fileCore)

		Logger = zap.New(core, zap.AddCaller())

	}

	return nil

}
