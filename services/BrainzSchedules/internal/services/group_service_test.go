package services

import (
	"brainz-api/internal/models"
	"context"
	"errors"
	"testing"
)

type mockGroupRepository struct {
	listFn                  func(ctx context.Context) ([]models.Group, error)
	getListForInstitutionFn func(ctx context.Context, instID int64) ([]models.Group, error)
}

func (m *mockGroupRepository) List(ctx context.Context) ([]models.Group, error) {
	return m.listFn(ctx)
}

func (m *mockGroupRepository) GetListForInstitution(ctx context.Context, instID int64) ([]models.Group, error) {
	return m.getListForInstitutionFn(ctx, instID)
}

func TestGroupService_GetGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &GroupService{
			gr: &mockGroupRepository{
				listFn: func(ctx context.Context) ([]models.Group, error) {
					return []models.Group{
						{ID: 1, Name: "A"},
						{ID: 2, Name: "B"},
					}, nil
				},
			},
		}

		got, err := svc.GetGroups(context.Background())
		if err != nil {
			t.Fatalf("GetGroups error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(groups) = %d, want 2", len(got))
		}
		if got[0].Id != 1 || got[0].Name != "A" {
			t.Fatalf("unexpected first group: %+v", got[0])
		}
	})

	t.Run("repository_error", func(t *testing.T) {
		wantErr := errors.New("list failed")
		svc := &GroupService{
			gr: &mockGroupRepository{
				listFn: func(ctx context.Context) ([]models.Group, error) {
					return nil, wantErr
				},
			},
		}

		_, err := svc.GetGroups(context.Background())
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})
}

func TestGroupService_GetGroupsForInstitution(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &GroupService{
			gr: &mockGroupRepository{
				getListForInstitutionFn: func(ctx context.Context, instID int64) ([]models.Group, error) {
					return []models.Group{
						{ID: 7, Name: "PI-101"},
						{ID: 8, Name: "PI-102"},
					}, nil
				},
			},
		}

		got, err := svc.GetGroupsForInstitution(context.Background(), 10)
		if err != nil {
			t.Fatalf("GetGroupsForInstitution error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(groups) = %d, want 2", len(got))
		}
		if got[1].Id != 8 || got[1].Name != "PI-102" {
			t.Fatalf("unexpected second group: %+v", got[1])
		}
	})

	t.Run("repository_error", func(t *testing.T) {
		wantErr := errors.New("institution groups failed")
		svc := &GroupService{
			gr: &mockGroupRepository{
				getListForInstitutionFn: func(ctx context.Context, instID int64) ([]models.Group, error) {
					return nil, wantErr
				},
			},
		}

		_, err := svc.GetGroupsForInstitution(context.Background(), 10)
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})
}
