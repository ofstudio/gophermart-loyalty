package logger

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"os"
	"strconv"
	"time"
)

type Log struct {
	zerolog.Logger
}

func NewLogger(level zerolog.Level) Log {

	l := zerolog.
		New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    false,
			TimeFormat: time.RFC3339,
		}).
		Level(level).
		With().Timestamp().Caller().
		Logger()

	return Log{Logger: l}
}

func (l Log) WithReqID(ctx context.Context) *Log {
	reqID := middleware.GetReqID(ctx)
	if reqID != "" {
		l.Logger = l.Logger.With().Str("request_id", reqID).Logger()
	}
	return &l
}

func init() {
	zerolog.CallerMarshalFunc = callerMarshalFunc
}

// callerMarshalFunc - возвращает имя файла и имя пакета в котором вызвана функция
func callerMarshalFunc(_ uintptr, filepath string, line int) string {
	slashCounter := 0
	for i := len(filepath) - 1; i >= 0; i-- {
		if filepath[i] == '/' {
			slashCounter++
		}
		if slashCounter == 2 {
			filepath = filepath[i+1:]
			break
		}
	}
	return filepath + ":" + strconv.Itoa(line)
}
