package chainlogger

import (
	"fmt"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// chainEncoder is a custom zap logger encoder that provides chain context to the logger
type chainEncoder struct {
	zapcore.Encoder
	pool  buffer.Pool
	chain string
	// id is unique number, the color of the chain depends on this value, two chains with the same id will have the same
	// color
	id int
}

// EncodeEntry is in charge of adding the chain context to each log
func (e *chainEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := e.pool.Get()

	prefix := fmt.Sprintf("[%s] ", e.chain)
	// Set color to chain name
	buf.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", e.id+31, prefix))

	consolebuf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(consolebuf.Bytes())
	if err != nil {
		return nil, err
	}
	return buf, nil
}
