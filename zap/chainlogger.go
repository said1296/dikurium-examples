package chainlogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"os"
)

// GetLogger generates a zap.Logger that provides chain context information to each log.
// The chain argument is prepended to each log, the id is necessary so as to generate a unique color to each chain log
// making readability easier.
func GetLogger(chain string, id int) *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := &chainEncoder{
		Encoder: zapcore.NewConsoleEncoder(config),
		pool:    buffer.NewPool(),
		chain:   chain,
		id:      id,
	}

	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
