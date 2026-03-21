package app

import (
	"brainz/developersapi/internal/config"
	"brainz/developersapi/internal/handler"
	"brainz/developersapi/internal/repository"
	"brainz/developersapi/internal/router"
	"brainz/developersapi/internal/security"
	"brainz/developersapi/internal/services"
	"brainz/developersapi/internal/storage"
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/cors"
	"go.uber.org/zap"
)

func sessionHTTPFromConfig(cfg *config.Config) handler.SessionHTTPConfig {
	return handler.NewSessionHTTPConfig(
		cfg.SessionCookieName,
		cfg.SessionCookieDomain,
		cfg.SessionCookiePath,
		cfg.SessionCookieMaxAge,
		cfg.SessionCookieSecure,
		parseCookieSameSite(cfg.SessionCookieSameSite),
	)
}

func parseCookieSameSite(s string) protocol.CookieSameSite {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "none":
		return protocol.CookieSameSiteNoneMode
	case "strict":
		return protocol.CookieSameSiteStrictMode
	case "lax", "":
		return protocol.CookieSameSiteLaxMode
	default:
		return protocol.CookieSameSiteLaxMode
	}
}

func corsFromConfig(cfg *config.Config) *cors.Config {
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowAllOrigins = false
	corsCfg.AllowCredentials = true
	corsCfg.AllowHeaders = []string{"*"}
	corsCfg.ExposeHeaders = []string{"*"}

	allowed := cfg.CorsOriginSet()
	if len(allowed) == 0 {
		// Dev: reflect any non-empty Origin (not "*"), required when the browser sends credentials.
		corsCfg.AllowOriginFunc = func(origin string) bool {
			return origin != ""
		}
	} else {
		corsCfg.AllowOriginFunc = func(origin string) bool {
			return allowed[origin]
		}
	}
	return &corsCfg
}

func Run(ctx context.Context, cfg *config.Config) error {
	hostPort := fmt.Sprintf("%s:%d", cfg.App.Address, cfg.App.Port)
	h := server.Default(server.WithHostPorts(hostPort))

	h.Use(cors.New(*corsFromConfig(cfg)))

	psqlPool, err := storage.ConnectPostgres(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}

	security.LoadCfg(&cfg.Auth)

	hc, err := client.NewClient()
	if err != nil {
		return err
	}

	sr := repository.NewSessionRepository(psqlPool)
	ur := repository.NewUserRepository(psqlPool)

	ss := services.NewSessionService(sr)
	us := services.NewUserService(ur)
	aks := services.NewApiKeysService(hc, "http://brainz-auth:8080", "")

	sc := sessionHTTPFromConfig(cfg)
	ah := handler.NewAuthHandler(us, ss, sc)
	kmh := handler.NewKeysManagementHandler(aks, ss, us, sc)

	router.Register(h, ah, kmh)

	h.OnShutdown = append(h.OnShutdown, func(ctx context.Context) {
		zap.L().Info("Stopping Server gracefully...")
		psqlPool.Close()
		<-ctx.Done()
		zap.L().Info("Server Stopped gracefully!")
	})

	h.Spin()
	return nil
}
