package main

import (
	"brainz/developersapi/internal/app"
	"brainz/developersapi/internal/config"
	"context"

	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
}

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := app.Run(ctx, nil, cfg); err != nil {
		panic(err)
	}
}
