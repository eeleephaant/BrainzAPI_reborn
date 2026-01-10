package repositories

import (
	"context"
	"errors"

	"brainz-api/internal/models"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepository struct {
	db *pgxpool.Pool
}

func NewGroupRepository(db *pgxpool.Pool) *GroupRepository {
	return &GroupRepository{db}
}

func (r *GroupRepository) GetListForInstitution(ctx context.Context, instID int64) ([]models.Group, error) {
	query, args, err := sq.
		Select("id", "created_at", "updated_at", "deleted_at", "name", "institution_id").
		From("groups").
		Where(sq.Eq{"institution_id": instID}).
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

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(
			&group.ID,
			&group.CreatedAt,
			&group.UpdatedAt,
			&group.DeletedAt,
			&group.Name,
			&group.InstitutionID,
		); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (r *GroupRepository) Create(ctx context.Context, group *models.Group) (int64, error) {
	query, args, err := sq.
		Insert("groups").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "institution_id").
		Values(group.Name, group.InstitutionID).
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

func (r *GroupRepository) GetByID(ctx context.Context, id int64) (*models.Group, error) {
	query, args, err := sq.
		Select("id", "created_at", "updated_at", "deleted_at", "name", "institution_id").
		From("groups").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	var group models.Group
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&group.ID,
		&group.CreatedAt,
		&group.UpdatedAt,
		&group.DeletedAt,
		&group.Name,
		&group.InstitutionID,
	)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) Update(ctx context.Context, group *models.Group) error {
	query, args, err := sq.
		Update("groups").
		Set("name", group.Name).
		Set("institution_id", group.InstitutionID).
		Set("updated_at", group.UpdatedAt).
		Where(sq.Eq{"id": group.ID}).
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
		return errors.New("group not found")
	}
	return nil
}

func (r *GroupRepository) Delete(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("groups").
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
		return errors.New("group not found")
	}
	return nil
}

func (r *GroupRepository) List(ctx context.Context) ([]models.Group, error) {
	query, args, err := sq.
		Select("id", "created_at", "updated_at", "deleted_at", "name", "institution_id").
		From("groups").
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

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(
			&group.ID,
			&group.CreatedAt,
			&group.UpdatedAt,
			&group.DeletedAt,
			&group.Name,
			&group.InstitutionID,
		); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}
