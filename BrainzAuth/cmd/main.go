package main

import (
	"brainz/auth/internal/app"
	"brainz/auth/internal/config"
	"context"

	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := app.Run(ctx, cfg); err != nil {
		panic(err)
	}
}
