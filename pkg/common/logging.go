package common

import (
	"go.uber.org/zap"
	"os"
)

var BaseLogger *zap.Logger
var Logger *zap.SugaredLogger

func init() {
	if os.Getenv("APP_ENV") == "production" {
		BaseLogger, _ = zap.NewProduction()
	} else {
		BaseLogger, _ = zap.NewDevelopment()
	}

	Logger = BaseLogger.Sugar()
}
