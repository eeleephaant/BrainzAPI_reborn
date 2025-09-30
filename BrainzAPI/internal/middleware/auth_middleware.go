package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"brainz-api/internal/models/dtos"

	"github.com/cloudwego/hertz/pkg/app"
)

func AuthMiddleware(verifyURL string) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		apiKey := string(ctx.Request.Header.Peek("X-API-Key"))
		if apiKey == "" {
			ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "missing api key"})
			ctx.Abort()
			return
		}

		client := &http.Client{Timeout: 3 * time.Second}
		req, err := http.NewRequestWithContext(
			c, http.MethodGet, fmt.Sprintf("%s?key=%s", verifyURL, url.QueryEscape(apiKey)), nil,
		)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
			ctx.Abort()
			return
		}

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid api key"})
			ctx.Abort()
			return
		}
		defer resp.Body.Close()

		var authResp dtos.AuthResult

		if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil || !authResp.Status {
			ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid api key"})
			ctx.Abort()
			return
		}

		ctx.Next(c)
	}
}
