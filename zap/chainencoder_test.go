package chainlogger

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"testing"
	"time"
)

func TestChainEncoder_EncodeEntry(t *testing.T) {
	id := 1
	expectedColorInt := 31 + id
	expectedChain := "eth_main"

	consoleEncoder := &chainEncoder{
		Encoder: zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
		pool:    buffer.NewPool(),
		chain:   expectedChain,
		id:      id,
	}

	zeroTime := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	entry := zapcore.Entry{
		LoggerName: "main",
		Level:      zapcore.InfoLevel,
		Message:    "hello",
		Time:       zeroTime,
		Stack:      "fake-stack",
		Caller:     zapcore.EntryCaller{Defined: true, File: "foo.go", Line: 42, Function: "foo.Foo"},
	}

	consoleOut, consoleErr := consoleEncoder.EncodeEntry(entry, nil)
	if assert.NoError(t, consoleErr, "unexpected error console-encoding entry") {
		assert.Equal(
			t,
			fmt.Sprintf("\x1b[%dm[%s] \x1b[0m0\tinfo\tmain\tfoo.go:42\thello\nfake-stack\n", expectedColorInt, expectedChain),
			consoleOut.String(),
			"unexpected console output: expected output to be prepended by chain context",
		)
	}
}
