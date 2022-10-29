package logger

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type Log struct {
	zerolog.Logger
}

func (l Log) Print(v ...interface{}) {
	l.Logger.Info().Msgf("%v", v)
}

func NewLogger(level zerolog.Level) Log {
	// Базовый логгер
	l := zerolog.
		New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    false,
			TimeFormat: time.RFC3339,
		}).
		Level(level).With().Timestamp().Logger()

	// Логирование для HTTP-запросов через chi middleware.Logger
	middleware.DefaultLogger = middleware.RequestLogger(
		&middleware.DefaultLogFormatter{
			Logger:  &Log{Logger: l},
			NoColor: false,
		})

	// Для остального логирования используем также Caller
	return Log{Logger: l.With().Caller().Logger()}
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
