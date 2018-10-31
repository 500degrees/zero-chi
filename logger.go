package zerochi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger interface {
	NewLogEntry(r *http.Request) middleware.LogEntry
}

func NewLogger() *StructuredLogger {
	return &StructuredLogger{}
}

func NewStructuredLogger(logger Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(logger)
}

type StructuredLogger struct {
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: log.Log()}

	logFields := map[string]interface{}{}

	logFields["ts"] = time.Now().UTC().Format(time.RFC1123)

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logFields["http_scheme"] = scheme
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	entry.Logger = entry.Logger.Fields(logFields)

	entry.Logger.Msg("request started")

	return entry
}

type StructuredLoggerEntry struct {
	Logger *zerolog.Event
}

func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.Fields(map[string]interface{}{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elapsed_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
	})

	l.Logger.Msg("request complete")
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.Fields(map[string]interface{}{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
	l.Logger.Msg("request failed")
}
