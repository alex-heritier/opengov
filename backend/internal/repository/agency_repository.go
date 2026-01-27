package repository

import (
	"context"
	"fmt"

	"github.com/alex/opengov-go/internal/db"
	"github.com/alex/opengov-go/internal/models"
)

type AgencyRepository struct {
	db *db.DB
}

func NewAgencyRepository(db *db.DB) *AgencyRepository {
	return &AgencyRepository{db: db}
}

func (r *AgencyRepository) GetAll(ctx context.Context, limit, offset int) ([]models.Agency, int, error) {
	query := "SELECT COUNT(*) FROM agencies"
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count agencies: %w", err)
	}

	query = `
		SELECT id, fr_agency_id, raw_name, name, short_name, slug, description, url, json_url, parent_id, raw_data, created_at, updated_at
		FROM agencies
		ORDER BY name
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query agencies: %w", err)
	}
	defer rows.Close()

	var agencies []models.Agency
	for rows.Next() {
		var a models.Agency
		var shortName, description, url, jsonURL *string
		var parentID *int
		if err := rows.Scan(
			&a.ID, &a.FRAgencyID, &a.RawName, &a.Name, &shortName, &a.Slug, &description,
			&url, &jsonURL, &parentID, &a.RawData, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan agency: %w", err)
		}
		a.ShortName = shortName
		a.Description = description
		a.URL = url
		a.JSONURL = jsonURL
		a.ParentID = parentID
		agencies = append(agencies, a)
	}

	return agencies, total, nil
}

func (r *AgencyRepository) Create(ctx context.Context, agency *models.Agency) error {
	query := `
		INSERT INTO agencies (fr_agency_id, raw_name, name, short_name, slug, description, url, json_url, parent_id, raw_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		agency.FRAgencyID, agency.RawName, agency.Name, agency.ShortName, agency.Slug,
		agency.Description, agency.URL, agency.JSONURL, agency.ParentID,
		agency.RawData,
	).Scan(&agency.ID)
	if err != nil {
		return fmt.Errorf("failed to insert agency: %w", err)
	}
	return nil
}

func (r *AgencyRepository) ExistsByFRAgencyID(ctx context.Context, frAgencyID int) (bool, error) {
	query := "SELECT COUNT(*) FROM agencies WHERE fr_agency_id = $1"
	var count int
	err := r.db.QueryRowContext(ctx, query, frAgencyID).Scan(&count)
	return count > 0, err
}

func (r *AgencyRepository) Upsert(ctx context.Context, agency *models.Agency) error {
	exists, err := r.ExistsByFRAgencyID(ctx, agency.FRAgencyID)
	if err != nil {
		return err
	}

	if exists {
		query := `
			UPDATE agencies SET raw_name=$1, name=$2, short_name=$3, slug=$4, description=$5, url=$6, json_url=$7, parent_id=$8, raw_data=$9, updated_at=NOW()
			WHERE fr_agency_id=$10
		`
		_, err = r.db.ExecContext(ctx, query,
			agency.RawName, agency.Name, agency.ShortName, agency.Slug, agency.Description,
			agency.URL, agency.JSONURL, agency.ParentID, agency.RawData,
			agency.FRAgencyID,
		)
		return err
	}

	return r.Create(ctx, agency)
}
