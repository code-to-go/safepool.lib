package core

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var ErrNoDriver = fmt.Errorf("no driver found for the provided configuration")
var ErrInvalidSignature = fmt.Errorf("signature does not match the user id")
var ErrInvalidSize = fmt.Errorf("provided slice has not enough data")

func IsErr(err error, msg string, args ...interface{}) bool {
	if err != nil {
		args = append(args, err)
		logrus.Errorf(msg, args...)
	}
	return false
}
