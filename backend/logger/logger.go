package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func GetLogger() *logrus.Logger {
	if Logger == nil {
		Logger = logrus.New()
		Logger.SetOutput(os.Stdout)
	}

	return Logger
}
