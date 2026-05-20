package logger

import (
	"go.uber.org/zap"
)

func InitLogger(logdev bool) (*zap.Logger, error) {
	var newlogger *zap.Logger
	var err error
	if logdev {
		newlogger, err = zap.NewDevelopment()
	} else {
		newlogger, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}
	return newlogger, nil
}
