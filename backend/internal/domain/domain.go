package domain

import (
	"time"

	"github.com/alex/opengov-go/internal/db/dbtypes"
)

// User is a DB-backed account record.
// Domain models MUST NOT contain transport tags (json, form, etc).
type User struct {
	ID               int64
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
	ID          int64
	FRAgencyID  int64
	RawName     string
	Name        string
	ShortName   *string
	Slug        string
	Description *string
	URL         *string
	JSONURL     *string
	ParentID    *int64
	RawData     dbtypes.JSONMap
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PolicyDocument struct {
	ID             int64
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
	ID          int64
	UserID      int64
	FeedEntryID int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Like struct {
	ID          int64
	UserID      int64
	FeedEntryID int64
	Value       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RawPolicyDocument struct {
	ID               int64
	SourceKey        string
	ExternalID       string
	RawData          dbtypes.JSONMap
	FetchedAt        time.Time
	PolicyDocumentID *int64
	CreatedAt        time.Time
}
