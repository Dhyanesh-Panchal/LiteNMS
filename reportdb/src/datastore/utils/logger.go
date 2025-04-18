package utils

import "go.uber.org/zap"

var Logger *zap.Logger

func init() {

	Logger = zap.Must(zap.NewDevelopment()) // New development for current basis

}
