package core

import (
	"github.com/sirupsen/logrus"
)

func Info(format string, args ...any) {
	logrus.Infof(format, args...)
}
