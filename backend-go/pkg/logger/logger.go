package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func resolveLogPath() string {
	if os.Getenv("TAURI_SIDEARCAR") == "1" {
		configDir, err := os.UserConfigDir()
		if err == nil {
			dir := filepath.Join(configDir, "投资助手")
			os.MkdirAll(dir, 0755)
			return filepath.Join(dir, "server.log")
		}
	}
	return "../data/server.log"
}

func Init() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	level := zapcore.InfoLevel
	if os.Getenv("DEBUG") == "1" {
		level = zapcore.DebugLevel
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   resolveLogPath(),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	})
	fileEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, level)

	core := zapcore.NewTee(consoleCore, fileCore)
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
