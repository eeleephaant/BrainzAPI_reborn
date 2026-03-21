package handler

import (
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
)

const defaultSessionCookieName = "brainz_session"
const defaultSessionMaxAgeSec = 14 * 24 * 3600 // matches SessionService session lifetime

// SessionHTTPConfig controls how the session token is stored in an HttpOnly cookie and read back.
type SessionHTTPConfig struct {
	CookieName   string
	CookieDomain string
	CookiePath   string
	MaxAgeSec    int
	Secure       bool
	SameSite     protocol.CookieSameSite
}

// NewSessionHTTPConfig builds settings with defaults for empty name/path/maxAge/sameSite.
func NewSessionHTTPConfig(name, domain, path string, maxAge int, secure bool, sameSite protocol.CookieSameSite) SessionHTTPConfig {
	if name == "" {
		name = defaultSessionCookieName
	}
	if path == "" {
		path = "/"
	}
	if maxAge <= 0 {
		maxAge = defaultSessionMaxAgeSec
	}
	if sameSite == protocol.CookieSameSiteDisabled {
		sameSite = protocol.CookieSameSiteLaxMode
	}
	return SessionHTTPConfig{
		CookieName:   name,
		CookieDomain: domain,
		CookiePath:   path,
		MaxAgeSec:    maxAge,
		Secure:       secure,
		SameSite:     sameSite,
	}
}

func (s SessionHTTPConfig) cookieName() string {
	if s.CookieName != "" {
		return s.CookieName
	}
	return defaultSessionCookieName
}

func (s SessionHTTPConfig) pathOrDefault() string {
	if s.CookiePath != "" {
		return s.CookiePath
	}
	return "/"
}

func (s SessionHTTPConfig) maxAgeOrDefault() int {
	if s.MaxAgeSec > 0 {
		return s.MaxAgeSec
	}
	return defaultSessionMaxAgeSec
}

func (s SessionHTTPConfig) sameSiteOrDefault() protocol.CookieSameSite {
	if s.SameSite != protocol.CookieSameSiteDisabled {
		return s.SameSite
	}
	return protocol.CookieSameSiteLaxMode
}

// TokenFromRequest returns the session token from X-Session-Token if set, otherwise from the session cookie.
func (s SessionHTTPConfig) TokenFromRequest(c *app.RequestContext) string {
	if t := strings.TrimSpace(c.Request.Header.Get("X-Session-Token")); t != "" {
		return t
	}
	b := c.Cookie(s.cookieName())
	if len(b) == 0 {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// SetSessionCookie sets HttpOnly cookie with the opaque session token (same value as JSON login response).
func (s SessionHTTPConfig) SetSessionCookie(c *app.RequestContext, token string) {
	c.SetCookie(s.cookieName(), token, s.maxAgeOrDefault(), s.pathOrDefault(), s.CookieDomain, s.sameSiteOrDefault(), s.Secure, true)
}
