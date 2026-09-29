package handler

import (
	"brainz-api/internal/dtos"
	"brainz/common/permissions"
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/ut"
)

type mockGroupService struct {
	getGroupsForInstitutionFn func(ctx context.Context, instID int64) ([]dtos.GroupDto, error)
}

func (m *mockGroupService) GetGroupsForInstitution(ctx context.Context, instID int64) ([]dtos.GroupDto, error) {
	return m.getGroupsForInstitutionFn(ctx, instID)
}

type mockAuthService struct {
	authorizeFn func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error)
}

func (m *mockAuthService) Authorize(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
	return m.authorizeFn(ctx, reqCtx, apiKey, perm)
}

func TestGroupHandler_GetGroupsForInst(t *testing.T) {
	ctx := context.Background()

	newCtx := func(target string) *app.RequestContext {
		return ut.CreateUtRequestContext("GET", target, nil)
	}

	t.Run("missing_institution_id", func(t *testing.T) {
		h := &GroupHandler{}
		c := newCtx("/group")
		h.GetGroupsForInst(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("invalid_institution_id", func(t *testing.T) {
		h := &GroupHandler{}
		c := newCtx("/group?institution_id=abc")
		h.GetGroupsForInst(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("authorize_error", func(t *testing.T) {
		h := &GroupHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, errors.New("auth failed")
				},
			},
		}
		c := newCtx("/group?institution_id=1")
		h.GetGroupsForInst(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		h := &GroupHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, nil
				},
			},
		}
		c := newCtx("/group?institution_id=1")
		h.GetGroupsForInst(ctx, c)
		if c.Response.StatusCode() != 403 {
			t.Fatalf("status = %d, want 403", c.Response.StatusCode())
		}
	})

	t.Run("service_error", func(t *testing.T) {
		h := &GroupHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			gs: &mockGroupService{
				getGroupsForInstitutionFn: func(ctx context.Context, instID int64) ([]dtos.GroupDto, error) {
					return nil, errors.New("db failed")
				},
			},
		}
		c := newCtx("/group?institution_id=1")
		h.GetGroupsForInst(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("empty_groups", func(t *testing.T) {
		h := &GroupHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			gs: &mockGroupService{
				getGroupsForInstitutionFn: func(ctx context.Context, instID int64) ([]dtos.GroupDto, error) {
					return []dtos.GroupDto{}, nil
				},
			},
		}
		c := newCtx("/group?institution_id=1")
		h.GetGroupsForInst(ctx, c)
		if c.Response.StatusCode() != 404 {
			t.Fatalf("status = %d, want 404", c.Response.StatusCode())
		}
	})

	t.Run("success", func(t *testing.T) {
		h := &GroupHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			gs: &mockGroupService{
				getGroupsForInstitutionFn: func(ctx context.Context, instID int64) ([]dtos.GroupDto, error) {
					return []dtos.GroupDto{{Id: 1, Name: "PI-101"}}, nil
				},
			},
		}
		c := newCtx("/group?institution_id=1")
		h.GetGroupsForInst(ctx, c)
		if c.Response.StatusCode() != 200 {
			t.Fatalf("status = %d, want 200", c.Response.StatusCode())
		}
		if !bytes.Contains(c.Response.Body(), []byte("PI-101")) {
			t.Fatalf("body = %s, want PI-101", c.Response.Body())
		}
	})
}
