package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

type NullTime struct {
	sql.NullTime
}

func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}
	nt.Valid = true
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
	default:
		return errors.New("type assertion to time.Time failed")
	}
	return nil
}

type User struct {
	ID               int        `json:"id"`
	Email            string     `json:"email"`
	HashedPassword   string     `json:"-"`
	IsActive         int        `json:"is_active"`
	IsSuperuser      int        `json:"is_superuser"`
	IsVerified       int        `json:"is_verified"`
	GoogleID         *string    `json:"google_id,omitempty"`
	Name             *string    `json:"name,omitempty"`
	PictureURL       *string    `json:"picture_url,omitempty"`
	PoliticalLeaning *string    `json:"political_leaning,omitempty"`
	State            *string    `json:"state,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
}

func (u *User) GetIsActive() bool {
	return u.IsActive == 1
}

func (u *User) GetIsSuperuser() bool {
	return u.IsSuperuser == 1
}

func (u *User) GetIsVerified() bool {
	return u.IsVerified == 1
}

type Agency struct {
	ID          int       `json:"id"`
	FRAgencyID  int       `json:"fr_agency_id"`
	RawName     string    `json:"raw_name"`
	Name        string    `json:"name"`
	ShortName   *string   `json:"short_name,omitempty"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	URL         *string   `json:"url,omitempty"`
	JSONURL     *string   `json:"json_url,omitempty"`
	ParentID    *int      `json:"parent_id,omitempty"`
	RawData     JSONMap   `json:"raw_data"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FederalRegisterDocument struct {
	ID             int       `json:"id"`
	FeedEntryID    int       `json:"feed_entry_id"`
	Source         string    `json:"source"`
	SourceID       string    `json:"source_id"`
	DocumentNumber string    `json:"document_number"`
	UniqueKey      string    `json:"unique_key"`
	RawData        JSONMap   `json:"raw_data"`
	FetchedAt      time.Time `json:"fetched_at"`
	Title          string    `json:"title"`
	Agency         *string   `json:"agency,omitempty"`
	Summary        string    `json:"summary"`
	Keypoints      []string  `json:"keypoints,omitempty"`
	ImpactScore    *string   `json:"impact_score,omitempty"`
	PoliticalScore *int      `json:"political_score,omitempty"`
	SourceURL      string    `json:"source_url"`
	PublishedAt    time.Time `json:"published_at"`
	DocumentType   *string   `json:"document_type,omitempty"`
	PDFURL         *string   `json:"pdf_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Bookmark struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	FeedEntryID int       `json:"feed_entry_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Like struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	FeedEntryID int       `json:"feed_entry_id"`
	Value       int       `json:"value"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
