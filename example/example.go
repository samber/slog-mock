package main

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	slogmock "github.com/samber/slog-mock"
)

func main() {
	logger := slog.New(
		slogmock.Option{
			Enabled: func(ctx context.Context, level slog.Level) bool {
				return true
			},
			Handle: func(ctx context.Context, record slog.Record) error {
				fmt.Printf("record: %s\n", record.Message)
				record.Attrs(func(attr slog.Attr) bool {
					fmt.Printf("   attr: %s\n", attr.Key)
					return true
				})
				return nil
			},
		}.NewMockHandler(),
		// slog.NewJSONHandler(os.Stdout, nil),
	)
	logger = logger.With("release", "v1.0.0")

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now()),
			),
		).
		With("error", fmt.Errorf("an error")).
		WithGroup("plop").
		Error("a message", slog.Int("count", 1))

	time.Sleep(1 * time.Second)
}
