package domain

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONMap is a convenience type for persisting arbitrary JSON payloads to Postgres.
// It implements database/sql interfaces for scanning and writing.
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

// NullTime mirrors sql.NullTime with a Scan that accepts time.Time.
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

// User is a DB-backed account record.
// Domain models MUST NOT contain transport tags (json, form, etc).
type User struct {
	ID               int
	Email            string
	HashedPassword   string
	IsActive         int
	IsSuperuser      int
	IsVerified       int
	GoogleID         *string
	Name             *string
	PictureURL       *string
	PoliticalLeaning *string
	State            *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastLoginAt      *time.Time
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
	ID          int
	FRAgencyID  int
	RawName     string
	Name        string
	ShortName   *string
	Slug        string
	Description *string
	URL         *string
	JSONURL     *string
	ParentID    *int
	RawData     JSONMap
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PolicyDocument struct {
	ID             int
	SourceKey      string
	ExternalID     string
	FetchedAt      time.Time
	Title          string
	Agency         *string
	Summary        string
	Keypoints      []string
	ImpactScore    *string
	PoliticalScore *int
	SourceURL      string
	PublishedAt    time.Time
	DocumentType   *string
	PDFURL         *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Bookmark struct {
	ID          int
	UserID      int
	FeedEntryID int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Like struct {
	ID          int
	UserID      int
	FeedEntryID int
	Value       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RawPolicyDocument struct {
	ID               int
	SourceKey        string
	ExternalID       string
	RawData          JSONMap
	FetchedAt        time.Time
	PolicyDocumentID int
	CreatedAt        time.Time
}
