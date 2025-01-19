package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ttl256/euivator/internal/fetcher"
)

func main() {
	loggerOptions := new(slog.HandlerOptions)
	loggerOptions.Level = slog.LevelDebug
	logger := slog.New(slog.NewTextHandler(os.Stdout, loggerOptions))

	fetch := fetcher.New(fetcher.GetSources(), "assets_ignore", logger)

	ctx := context.Background()

	err := fetch.DownloadFiles(ctx)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "error", slog.Any("error", err))
		os.Exit(1)
	}

	// URL, err := url.Parse("https://test.com")
	// if err != nil {
	// 	logger.LogAttrs(ctx, slog.LevelError, "error", slog.Any("error", err))
	// 	os.Exit(1)
	// }

	// var m = map[string]string{
	// 	"https://test1.com": "test1",
	// 	"https://test2.com": "test2",
	// 	"https://test3.com": "test3",
	// 	"https://test4.com": "test4",
	// }

	// data, err := fetch.LoadETags()
	// if err != nil {
	// 	logger.LogAttrs(ctx, slog.LevelError, "error", slog.Any("error", err))
	// 	os.Exit(1)
	// }
	// fmt.Println(data)
	// data["https://test5.com"] = "test5"

	// err = fetch.SaveETags(data)
	// if err != nil {
	// 	logger.LogAttrs(ctx, slog.LevelError, "error", slog.Any("error", err))
	// 	os.Exit(1)
	// }

	logger.LogAttrs(ctx, slog.LevelInfo, "all done")
}
