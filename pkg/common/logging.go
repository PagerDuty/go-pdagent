package common

import (
	"go.uber.org/zap"
)

var BaseLogger *zap.Logger
var Logger *zap.SugaredLogger

// TODO: Eventually move configuration to config files.
func init() {
	if IsProduction() {
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{
			"/var/log/pdagent/pdagent.log",
		}
		BaseLogger, _ = config.Build()
	} else {
		BaseLogger, _ = zap.NewDevelopment()
	}

	Logger = BaseLogger.Sugar()
}
