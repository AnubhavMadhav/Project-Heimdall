package logger

import (
	"log/slog"
	"os"
)

// New initializes a structured JSON logger.
func New(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		// AddSource: true, // Uncomment if you want file:line in production logs
	}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	return slog.New(handler)
}
