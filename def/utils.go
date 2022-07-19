package def

import "github.com/sirupsen/logrus"

func IsErr(err error, msg string, args ...interface{}) bool {
	if err != nil {
		args = append(args, err)
		logrus.Errorf(msg, args...)
	}
	return false
}
