package repository

import (
	"context"
	"database/sql"
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
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query agencies: %w", err)
	}
	defer rows.Close()

	var agencies []models.Agency
	for rows.Next() {
		var a models.Agency
		var shortName, description, url, jsonURL sql.NullString
		var parentID sql.NullInt64
		var createdAt, updatedAt string
		if err := rows.Scan(
			&a.ID, &a.FRAgencyID, &a.RawName, &a.Name, &shortName, &a.Slug, &description,
			&url, &jsonURL, &parentID, &a.RawData, &createdAt, &updatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan agency: %w", err)
		}
		if shortName.Valid {
			a.ShortName = &shortName.String
		}
		if description.Valid {
			a.Description = &description.String
		}
		if url.Valid {
			a.URL = &url.String
		}
		if jsonURL.Valid {
			a.JSONURL = &jsonURL.String
		}
		if parentID.Valid {
			pid := int(parentID.Int64)
			a.ParentID = &pid
		}
		a.CreatedAt = createdAt
		a.UpdatedAt = updatedAt
		agencies = append(agencies, a)
	}

	return agencies, total, nil
}

func (r *AgencyRepository) Create(ctx context.Context, agency *models.Agency) error {
	query := `
		INSERT INTO agencies (fr_agency_id, raw_name, name, short_name, slug, description, url, json_url, parent_id, raw_data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		agency.FRAgencyID, agency.RawName, agency.Name, agency.ShortName, agency.Slug,
		agency.Description, agency.URL, agency.JSONURL, agency.ParentID,
		agency.RawData, agency.CreatedAt, agency.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert agency: %w", err)
	}
	id, _ := result.LastInsertId()
	agency.ID = int(id)
	return nil
}

func (r *AgencyRepository) ExistsByFRAgencyID(ctx context.Context, frAgencyID int) (bool, error) {
	query := "SELECT COUNT(*) FROM agencies WHERE fr_agency_id = ?"
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
			UPDATE agencies SET raw_name=?, name=?, short_name=?, slug=?, description=?, url=?, json_url=?, parent_id=?, raw_data=?, updated_at=?
			WHERE fr_agency_id=?
		`
		_, err = r.db.ExecContext(ctx, query,
			agency.RawName, agency.Name, agency.ShortName, agency.Slug, agency.Description,
			agency.URL, agency.JSONURL, agency.ParentID, agency.RawData,
			agency.UpdatedAt, agency.FRAgencyID,
		)
		return err
	}

	return r.Create(ctx, agency)
}
