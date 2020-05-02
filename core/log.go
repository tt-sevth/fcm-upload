/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: log.go
 * Date: 2020/5/2 上午1:19
 * Author: sevth
 */

package core

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

// 初始化日志模块
func NewLogger() *zap.SugaredLogger {
	writeSync := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSync, zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	sugarLogger := logger.Sugar()
	return sugarLogger
}

// 日志写入文件
func getLogWriter() zapcore.WriteSyncer {
	if _, err := os.Stat(getHomeDir() + "/FCM/log"); err != nil {
		_ = os.MkdirAll(getHomeDir() + "/FCM/log", 0755)
	}
	filePath := getHomeDir() + "/FCM/log/" + time.Now().Format("2006-01-02") + ".log"
	logFile, err := os.OpenFile(filePath,  os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(logFile)
}

// 一些编码设置
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}
