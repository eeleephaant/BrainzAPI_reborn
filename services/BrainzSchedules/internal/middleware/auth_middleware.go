package middleware

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"brainz-api/internal/db"
	"brainz-api/internal/dtos"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(verifyURL string, client *http.Client) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		apiKey := string(ctx.Request.Header.Peek("X-API-Key"))
		if apiKey == "" {
			abort(ctx, http.StatusUnauthorized, "missing api key")
			return
		}

		cacheKey := "api_key:" + apiKey

		_, err := db.RedisClient.GetEx(c, cacheKey, 2*time.Minute).Result()
		if err == nil {
			ctx.Next(c)
			return
		}
		if err != redis.Nil {
			abort(ctx, http.StatusInternalServerError, "redis error")
			return
		}

		u, _ := url.Parse(verifyURL)
		q := u.Query()
		q.Set("key", apiKey)
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(c, http.MethodGet, u.String(), nil)
		if err != nil {
			abort(ctx, http.StatusInternalServerError, "internal error")
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			abort(ctx, http.StatusUnauthorized, "invalid api key")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			abort(ctx, http.StatusUnauthorized, "invalid api key")
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			abort(ctx, http.StatusUnauthorized, "invalid api key")
			return
		}

		var authResp dtos.AuthResult
		if err := sonic.Unmarshal(body, &authResp); err != nil || !authResp.Status {
			abort(ctx, http.StatusUnauthorized, "invalid api key")
			return
		}

		db.RedisClient.Set(c, cacheKey, "valid", 5*time.Minute)

		ctx.Next(c)
	}
}

func abort(ctx *app.RequestContext, code int, msg string) {
	ctx.JSON(code, map[string]string{"error": msg})
	ctx.Abort()
}
