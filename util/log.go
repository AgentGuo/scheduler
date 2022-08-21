package util

import (
	"context"
	"fmt"
	"github.com/AgentGuo/scheduler/cmd/scheduler-main/config"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

const (
	LoggerKey = "__LOGGER__"
	FieldsKey = "__FIELDS__"
)

var logLevelMap = map[string]logrus.Level{
	"TraceLevel": logrus.TraceLevel,
	"DebugLevel": logrus.DebugLevel,
	"InfoLevel":  logrus.InfoLevel,
	"WarnLevel":  logrus.WarnLevel,
	"ErrorLevel": logrus.ErrorLevel,
	"FatalLevel": logrus.FatalLevel,
	"PanicLevel": logrus.PanicLevel,
}

// InitLog 初始化log
func InitLog(config *config.SchedulerMainConfig) *logrus.Logger {
	logger := logrus.New()
	logger.Out = os.Stdout
	logger.SetFormatter(&logrus.TextFormatter{}) // 日志格式
	logger.SetReportCaller(false)                // 打印调用方法
	if level, ok := logLevelMap[config.LogLevel]; ok {
		logger.SetLevel(level)
	} else {
		logger.SetLevel(logrus.WarnLevel) // 默认设置为WarnLevel
		logger.Warnf("level set failed, config level:%+v", config.LogLevel)
	}
	return logger
}

func SetCtxLogger(ctx context.Context, logger *logrus.Logger) context.Context {
	ctx = context.WithValue(ctx, LoggerKey, logger)
	fieldsMap := &sync.Map{}
	ctx = context.WithValue(ctx, FieldsKey, fieldsMap)
	return ctx
}

func GetCtxLogger(ctx context.Context) (*logrus.Entry, error) {
	v := ctx.Value(LoggerKey)
	if logger, ok := v.(*logrus.Logger); ok {
		v := ctx.Value(FieldsKey)
		if fieldsMap, ok := v.(*sync.Map); ok {
			fieldList := logrus.Fields{}
			fieldsMap.Range(func(key, value any) bool {
				fieldList[key.(string)] = value
				return true
			})
			return logger.WithFields(fieldList), nil
		} else {
			return nil, fmt.Errorf("get fields failed")
		}
	} else {
		return nil, fmt.Errorf("get logger failed")
	}
}

func SetCtxFields(ctx context.Context, fields map[string]string) {
	v := ctx.Value(FieldsKey)
	if fieldsMap, ok := v.(*sync.Map); ok {
		for k, v := range fields {
			fieldsMap.Store(k, v)
		}
	}
}
