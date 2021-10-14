package anxcloud

import (
	"fmt"
	"log"

	"github.com/go-logr/logr"
)

var logger logr.Logger

func setupLogger() {
	logger = NewTerraformr(log.Default().Writer())
}

func LogDebug(msg string, args ...interface{}) {
	logger.V(2).Info(fmt.Sprintf(msg, args...))
}

func LogInfo(msg string, args ...interface{}) {
	logger.V(1).Info(fmt.Sprintf(msg, args...))
}

func LogWarn(msg string, args ...interface{}) {
	logger.Info(fmt.Sprintf(msg, args...))
}

func LogError(msg string, args ...interface{}) {
	logger.Error(nil, fmt.Sprintf(msg, args...))
}
