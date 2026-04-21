package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	sugar        *zap.SugaredLogger
)

// Init 初始化日志
func Init(level, format, outputPath string) error {
	zapLevel := parseLevel(level)

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if format == "console" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if outputPath != "" {
		config.OutputPaths = []string{outputPath, "stdout"}
		config.ErrorOutputPaths = []string{outputPath, "stderr"}
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	globalLogger = logger
	sugar = logger.Sugar()

	return nil
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Logger 获取 zap Logger
func Logger() *zap.Logger {
	if globalLogger == nil {
		// 默认初始化
		globalLogger, _ = zap.NewProduction()
		sugar = globalLogger.Sugar()
	}
	return globalLogger
}

// Sugar 获取 sugared logger
func Sugar() *zap.SugaredLogger {
	if sugar == nil {
		Logger()
	}
	return sugar
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	Logger().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	Logger().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	Logger().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	Logger().Error(msg, fields...)
}

// Fatal 致命错误
func Fatal(msg string, fields ...zap.Field) {
	Logger().Fatal(msg, fields...)
	os.Exit(1)
}

// Debugf 调试日志（格式化）
func Debugf(template string, args ...interface{}) {
	Sugar().Debugf(template, args...)
}

// Infof 信息日志（格式化）
func Infof(template string, args ...interface{}) {
	Sugar().Infof(template, args...)
}

// Warnf 警告日志（格式化）
func Warnf(template string, args ...interface{}) {
	Sugar().Warnf(template, args...)
}

// Errorf 错误日志（格式化）
func Errorf(template string, args ...interface{}) {
	Sugar().Errorf(template, args...)
}

// Fatalf 致命错误（格式化）
func Fatalf(template string, args ...interface{}) {
	Sugar().Fatalf(template, args...)
	os.Exit(1)
}

// With 添加上下文
func With(fields ...zap.Field) *zap.Logger {
	return Logger().With(fields...)
}

// Sync 同步日志
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
