package logger

import (
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
)

func InitializeLogger() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level: slog.LevelDebug,
		})),
	)
}
