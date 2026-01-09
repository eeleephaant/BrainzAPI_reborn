package repositories

import (
	"brainz-api/internal/models"
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InstitutionRepository struct {
	db *pgxpool.Pool
}

func NewInstitutionRepository(db *pgxpool.Pool) *InstitutionRepository {
	return &InstitutionRepository{db}
}

// Create inserts a new institution and returns its ID.
func (r *InstitutionRepository) Create(ctx context.Context, institution *models.Institution) (int64, error) {
	query, args, err := sq.
		Insert("institutions").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "site_link").
		Values(institution.Name, institution.SiteLink).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, err
	}
	var id int64
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID returns an institution by its ID.
func (r *InstitutionRepository) GetByID(ctx context.Context, id int64) (*models.Institution, error) {
	query, args, err := sq.
		Select("id", "created_at", "updated_at", "deleted_at", "name", "site_link").
		From("institutions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	var institution models.Institution
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&institution.ID,
		&institution.CreatedAt,
		&institution.UpdatedAt,
		&institution.DeletedAt,
		&institution.Name,
		&institution.SiteLink,
	)
	if err != nil {
		return nil, err
	}
	return &institution, nil
}

// Update updates institution fields by ID.
func (r *InstitutionRepository) Update(ctx context.Context, institution *models.Institution) error {
	query, args, err := sq.
		Update("institutions").
		Set("name", institution.Name).
		Set("site_link", institution.SiteLink).
		Set("updated_at", institution.UpdatedAt).
		Where(sq.Eq{"id": institution.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("institution not found")
	}
	return nil
}

// Delete removes an institution by ID.
func (r *InstitutionRepository) Delete(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("institutions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("institution not found")
	}
	return nil
}

// List returns all institutions.
func (r *InstitutionRepository) List(ctx context.Context) ([]*models.Institution, error) {
	query, args, err := sq.
		Select("id", "created_at", "updated_at", "deleted_at", "name", "site_link").
		From("institutions").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var institutions []*models.Institution
	for rows.Next() {
		var institution models.Institution
		if err := rows.Scan(
			&institution.ID,
			&institution.CreatedAt,
			&institution.UpdatedAt,
			&institution.DeletedAt,
			&institution.Name,
			&institution.SiteLink,
		); err != nil {
			return nil, err
		}
		institutions = append(institutions, &institution)
	}
	return institutions, nil
}
