package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.SugaredLogger
	mux    *sync.Mutex
}

var l *Logger

func CreateLogger() error {
	// set up logging
	zapLogger, err := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths: []string{"logs/logs.txt"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message", // <--
			NameKey:    "name",

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
		},
	}.Build()

	if err != nil {
		return err
	}

	l = &Logger{
		logger: zapLogger.Sugar(),
		mux:    &sync.Mutex{},
	}

	return nil
}

func Named(name string) *Logger {
	return &Logger{
		logger: l.logger.Named(name),
		mux:    &sync.Mutex{},
	}
}

func Debug(args ...interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.logger.Debug(args...)
}

func Info(args ...interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.logger.Info(args)
}

func Warn(args ...interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.logger.Warn(args)
}

func (l *Logger) Debug(args ...interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.logger.Debug(args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.logger.Info(args)
}

func (l *Logger) Warn(args ...interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.logger.Warn(args)
}
