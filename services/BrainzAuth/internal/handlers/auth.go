package handlers

import (
	"brainz/auth/internal/dtos"
	"brainz/auth/internal/services"
	"brainz/auth/internal/utils"
	cmnDtos "brainz/common/dtos"
	"context"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthHandler struct {
	aks  *services.ApiKeysService
	akus *services.ApiKeyUsageService
	rls  *services.RateLimiterService
}

func NewAuthHandler(aks *services.ApiKeysService, akus *services.ApiKeyUsageService, rls *services.RateLimiterService) *AuthHandler {
	return &AuthHandler{
		aks:  aks,
		akus: akus,
		rls:  rls,
	}
}

func (ah *AuthHandler) Auth(ctx context.Context, c *app.RequestContext) {
	ipAddr := extractClientIP(c)

	ipDecision, err := ah.allowRate(ctx, ipAddr, nil)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "rate limiter unavailable",
		})
		return
	}
	writeRateHeaders(c, ipDecision)
	if !ipDecision.Allowed {
		c.JSON(http.StatusTooManyRequests, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "rate limit exceeded",
		})
		zap.L().Warn("IP rate limit exceeded", zap.String("ip", ipAddr))
		return
	}

	apiKey := extractApiKey(c)
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "missing apiKey",
		})
		zap.L().Error("Missing apiKey in request")
		return
	}

	keyData, parseErr := utils.ExtractDataFromKey(apiKey)
	if parseErr == nil {
		keyID := keyData.ID
		keyDecision, keyLimitErr := ah.allowRate(ctx, ipAddr, &keyID)
		if keyLimitErr != nil {
			c.JSON(http.StatusServiceUnavailable, dtos.AuthResult{
				Status:       false,
				ErrorMessage: "rate limiter unavailable",
			})
			return
		}
		writeRateHeaders(c, keyDecision)
		if !keyDecision.Allowed {
			c.JSON(http.StatusTooManyRequests, dtos.AuthResult{
				Status:       false,
				ErrorMessage: "rate limit exceeded",
			})
			zap.L().Warn("API key rate limit exceeded", zap.String("ip", ipAddr), zap.String("api_key_id", keyData.ID.String()))
			return
		}
	}

	requestedPermissionDTO := cmnDtos.RequestedPermissionDTO{}
	if err := c.BindAndValidate(&requestedPermissionDTO); err != nil {
		c.String(400, err.Error())
		return
	}

	err = ah.aks.ValidateKey(ctx, apiKey, ipAddr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dtos.AuthResult{
			Status:       false,
			ErrorMessage: "invalid apiKey",
		})
		zap.L().Error("API key validation failed", zap.Error(err))
		return
	}
	isLegit, err := ah.aks.CheckPermission(ctx, apiKey, requestedPermissionDTO.Permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		zap.L().Error("API key checking permissions failed", zap.Error(err))
		return
	}
	if !isLegit {
		c.JSON(http.StatusForbidden, dtos.AuthResult{Status: false, ErrorMessage: "API key does not have permission"})
		zap.L().Error("API key does not have permission", zap.Any("requestedPermissionDTO", requestedPermissionDTO))
		return
	}
	c.JSON(http.StatusOK, dtos.AuthResult{
		Status:       true,
		ErrorMessage: "",
	})
	zap.L().Info("Auth successful")

	path := c.Request.Header.Get("X-Original-Path")
	method := c.Request.Header.Get("X-Original-Method")
	_, err = ah.akus.RecordUsage(ctx, apiKey, path, method)
	if err != nil {
		zap.L().Error("Usage record error: ", zap.Error(err))
		return
	}

}

func extractClientIP(c *app.RequestContext) string {
	if ip := strings.TrimSpace(c.Request.Header.Get("X-Original-IP")); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(c.Request.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	if forwarded := strings.TrimSpace(c.Request.Header.Get("X-Forwarded-For")); forwarded != "" {
		first, _, _ := strings.Cut(forwarded, ",")
		return strings.TrimSpace(first)
	}
	return c.ClientIP()
}

func writeRateHeaders(c *app.RequestContext, decision services.RateLimitDecision) {
	if decision.Limit > 0 {
		c.Response.Header.Set("X-RateLimit-Limit", strconv.FormatInt(decision.Limit, 10))
		c.Response.Header.Set("X-RateLimit-Remaining", strconv.FormatInt(decision.Remaining, 10))
	}
	if decision.ResetAfter > 0 {
		resetIn := int64(math.Ceil(decision.ResetAfter.Seconds()))
		if resetIn < 1 {
			resetIn = 1
		}
		c.Response.Header.Set("X-RateLimit-Reset", strconv.FormatInt(resetIn, 10))
	}
	if decision.RetryAfter > 0 {
		retryIn := int64(math.Ceil(decision.RetryAfter.Seconds()))
		if retryIn < 1 {
			retryIn = 1
		}
		c.Response.Header.Set("Retry-After", strconv.FormatInt(retryIn, 10))
	}
	if decision.Scope != "" {
		c.Response.Header.Set("X-RateLimit-Scope", decision.Scope)
	}
}

func (ah *AuthHandler) allowRate(ctx context.Context, ipAddr string, keyID *uuid.UUID) (services.RateLimitDecision, error) {
	if ah.rls == nil || !ah.rls.Enabled() {
		return services.RateLimitDecision{Allowed: true}, nil
	}

	var (
		decision services.RateLimitDecision
		err      error
	)
	if keyID == nil {
		decision, err = ah.rls.AllowIP(ctx, ipAddr)
	} else {
		decision, err = ah.rls.AllowKey(ctx, *keyID)
	}
	if err == nil {
		return decision, nil
	}
	if ah.rls.ShouldFailOpen() {
		apiKeyID := ""
		if keyID != nil {
			apiKeyID = keyID.String()
		}
		zap.L().Warn("rate limiter failed in fail-open mode", zap.Error(err), zap.String("scope", decision.Scope), zap.String("ip", ipAddr), zap.String("api_key_id", apiKeyID))
		return services.RateLimitDecision{Allowed: true}, nil
	}
	return services.RateLimitDecision{}, err
}

func extractApiKey(c *app.RequestContext) string {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader != "" {
		const bearerPrefix = "Bearer "
		if strings.HasPrefix(authHeader, bearerPrefix) {
			return strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		}
	}

	apiKey := c.Request.Header.Get("X-API-Key")
	if apiKey != "" {
		return strings.TrimSpace(apiKey)
	}

	apiKey = c.Request.Header.Get("X-Api-Key")
	if apiKey != "" {
		return strings.TrimSpace(apiKey)
	}

	return ""
}
