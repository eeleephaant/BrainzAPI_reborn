package handler

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/common/ut"
)

func TestSessionHTTPConfig_TokenFromRequest_cookie_query_unescape(t *testing.T) {
	sc := SessionHTTPConfig{}
	c := ut.CreateUtRequestContext("GET", "/x", nil)
	// Browser sends the cookie value as stored; Hertz wrote it with QueryEscape, so ":" is %3A.
	c.Request.Header.Set("Cookie", "brainz_session=part1%3Apart2")

	got := sc.TokenFromRequest(c)
	want := "part1:part2"
	if got != want {
		t.Fatalf("TokenFromRequest = %q, want %q", got, want)
	}
}

func TestSessionHTTPConfig_TokenFromRequest_header_not_unescaped(t *testing.T) {
	sc := SessionHTTPConfig{}
	c := ut.CreateUtRequestContext("GET", "/x", nil)
	c.Request.Header.Set("X-Session-Token", "literal%3Akeep")
	// Cookie would decode; header must stay raw for callers that send literal tokens.
	got := sc.TokenFromRequest(c)
	if got != "literal%3Akeep" {
		t.Fatalf("TokenFromRequest = %q, want literal %%3A unchanged from header", got)
	}
}
