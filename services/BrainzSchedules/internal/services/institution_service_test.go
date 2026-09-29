package services

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/models"
	"context"
	"errors"
	"testing"
)

type mockInstitutionRepository struct {
	createFn func(ctx context.Context, institution *models.Institution) (int64, error)
	listFn   func(ctx context.Context) ([]*models.Institution, error)
}

func (m *mockInstitutionRepository) Create(ctx context.Context, institution *models.Institution) (int64, error) {
	return m.createFn(ctx, institution)
}

func (m *mockInstitutionRepository) List(ctx context.Context) ([]*models.Institution, error) {
	return m.listFn(ctx)
}

func TestInstitutionService_CreateNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var captured *models.Institution
		svc := &InstitutionService{
			ir: &mockInstitutionRepository{
				createFn: func(ctx context.Context, institution *models.Institution) (int64, error) {
					captured = institution
					return 55, nil
				},
			},
		}

		got, err := svc.CreateNew(context.Background(), dtos.InstitutionCreateDto{Name: "Brainz", SiteLink: "https://brainz.dev"})
		if err != nil {
			t.Fatalf("CreateNew error = %v", err)
		}
		if got != 55 {
			t.Fatalf("created id = %d, want 55", got)
		}
		if captured == nil || captured.SiteLink == nil || *captured.SiteLink != "https://brainz.dev" {
			t.Fatalf("unexpected institution passed to repository: %+v", captured)
		}
	})

	t.Run("repository_error", func(t *testing.T) {
		wantErr := errors.New("create failed")
		svc := &InstitutionService{
			ir: &mockInstitutionRepository{
				createFn: func(ctx context.Context, institution *models.Institution) (int64, error) {
					return 0, wantErr
				},
			},
		}

		_, err := svc.CreateNew(context.Background(), dtos.InstitutionCreateDto{Name: "Brainz"})
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})
}

func TestInstitutionService_GetInstitutions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		site := "https://brainz.dev"
		svc := &InstitutionService{
			ir: &mockInstitutionRepository{
				listFn: func(ctx context.Context) ([]*models.Institution, error) {
					return []*models.Institution{
						{ID: 1, Name: "Brainz", SiteLink: &site},
						{ID: 2, Name: "NoSite", SiteLink: nil},
					}, nil
				},
			},
		}

		got, err := svc.GetInstitutions(context.Background())
		if err != nil {
			t.Fatalf("GetInstitutions error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(institutions) = %d, want 2", len(got))
		}
		if got[0].Site != site {
			t.Fatalf("site = %q, want %q", got[0].Site, site)
		}
		if got[1].Site != "" {
			t.Fatalf("site for nil link = %q, want empty string", got[1].Site)
		}
	})

	t.Run("repository_error", func(t *testing.T) {
		wantErr := errors.New("list failed")
		svc := &InstitutionService{
			ir: &mockInstitutionRepository{
				listFn: func(ctx context.Context) ([]*models.Institution, error) {
					return nil, wantErr
				},
			},
		}

		_, err := svc.GetInstitutions(context.Background())
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})
}
