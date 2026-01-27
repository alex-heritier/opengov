package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alex/opengov-go/internal/db"
)

type FeedRepository struct {
	db *db.DB
}

func NewFeedRepository(db *db.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

type FeedEntryRow struct {
	FeedEntryID int
	PublishedAt time.Time

	Title          string
	ShortText      string
	KeyPoints      []string
	PoliticalScore *int
	ImpactScore    *string
	SourceURL      string

	IsBookmarked   *bool
	UserLikeStatus *int
	LikesCount     int
	DislikesCount  int
}

func (r *FeedRepository) GetFeedAnon(ctx context.Context, page, limit int, sort string) ([]FeedEntryRow, int, error) {
	offset := (page - 1) * limit
	var orderDir string
	if sort == "newest" {
		orderDir = "DESC"
	} else {
		orderDir = "ASC"
	}

	fromWhere := "FROM feed_entries fi"
	whereClause := ""
	likesAggJoin := `
		LEFT JOIN (
			SELECT
				feed_entry_id,
				SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END) AS likes_count,
				SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END) AS dislikes_count
			FROM likes
			GROUP BY feed_entry_id
		) agg ON agg.feed_entry_id = fi.id
	`
	baseQuery := fmt.Sprintf("%s\n%s\n%s", fromWhere, likesAggJoin, whereClause)

	query := fmt.Sprintf(`
		SELECT
			fi.id AS feed_entry_id,
			fi.published_at,
			fi.title,
			fi.short_text,
			fi.key_points,
			fi.political_score,
			fi.impact_score,
			fi.source_url,
			COALESCE(agg.likes_count, 0) AS likes_count,
			COALESCE(agg.dislikes_count, 0) AS dislikes_count
		%s
		ORDER BY fi.published_at %s
		LIMIT $1 OFFSET $2
	`, baseQuery, orderDir)

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query feed: %w", err)
	}
	defer rows.Close()

	var items []FeedEntryRow
	for rows.Next() {
		var item FeedEntryRow
		var keyPointsRaw []byte
		var politicalScore sql.NullInt64
		var impactScore sql.NullString
		err := rows.Scan(
			&item.FeedEntryID,
			&item.PublishedAt,
			&item.Title,
			&item.ShortText,
			&keyPointsRaw,
			&politicalScore,
			&impactScore,
			&item.SourceURL,
			&item.LikesCount,
			&item.DislikesCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan feed entry: %w", err)
		}
		if politicalScore.Valid {
			ps := int(politicalScore.Int64)
			item.PoliticalScore = &ps
		}
		if impactScore.Valid {
			item.ImpactScore = &impactScore.String
		}
		if len(keyPointsRaw) > 0 {
			if err := json.Unmarshal(keyPointsRaw, &item.KeyPoints); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal key_points: %w", err)
			}
		}
		items = append(items, item)
	}

	var total int
	countQuery := "SELECT COUNT(DISTINCT fi.id)\n" + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count feed entrys: %w", err)
	}

	return items, total, nil
}

func (r *FeedRepository) GetFeedForUser(ctx context.Context, userID, page, limit int, sort string) ([]FeedEntryRow, int, error) {
	offset := (page - 1) * limit
	var orderDir string
	if sort == "newest" {
		orderDir = "DESC"
	} else {
		orderDir = "ASC"
	}

	fromWhere := "FROM feed_entries fi"
	whereClause := ""
	likesAggJoin := `
		LEFT JOIN (
			SELECT
				feed_entry_id,
				SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END) AS likes_count,
				SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END) AS dislikes_count
			FROM likes
			GROUP BY feed_entry_id
		) agg ON agg.feed_entry_id = fi.id
	`
	userJoin := `
		LEFT JOIN bookmarks b ON b.feed_entry_id = fi.id AND b.user_id = $1
		LEFT JOIN likes ul ON ul.feed_entry_id = fi.id AND ul.user_id = $1
	`
	baseQuery := fmt.Sprintf("%s\n%s\n%s\n%s", fromWhere, likesAggJoin, userJoin, whereClause)

	query := fmt.Sprintf(`
		SELECT
			fi.id AS feed_entry_id,
			fi.published_at,
			fi.title,
			fi.short_text,
			fi.key_points,
			fi.political_score,
			fi.impact_score,
			fi.source_url,
			COALESCE(agg.likes_count, 0) AS likes_count,
			COALESCE(agg.dislikes_count, 0) AS dislikes_count,
			(CASE WHEN b.feed_entry_id IS NULL THEN FALSE ELSE TRUE END) AS is_bookmarked,
			ul.value AS user_like_status
		%s
		ORDER BY fi.published_at %s
		LIMIT $2 OFFSET $3
	`, baseQuery, orderDir)

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query feed for user: %w", err)
	}
	defer rows.Close()

	var items []FeedEntryRow
	for rows.Next() {
		var item FeedEntryRow
		var keyPointsRaw []byte
		var politicalScore sql.NullInt64
		var impactScore sql.NullString
		var isBookmarked bool
		var userLikeStatus sql.NullInt64
		err := rows.Scan(
			&item.FeedEntryID,
			&item.PublishedAt,
			&item.Title,
			&item.ShortText,
			&keyPointsRaw,
			&politicalScore,
			&impactScore,
			&item.SourceURL,
			&item.LikesCount,
			&item.DislikesCount,
			&isBookmarked,
			&userLikeStatus,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan feed entry: %w", err)
		}
		if politicalScore.Valid {
			ps := int(politicalScore.Int64)
			item.PoliticalScore = &ps
		}
		if impactScore.Valid {
			item.ImpactScore = &impactScore.String
		}
		bookmarked := isBookmarked
		item.IsBookmarked = &bookmarked
		if userLikeStatus.Valid {
			uls := int(userLikeStatus.Int64)
			item.UserLikeStatus = &uls
		}
		if len(keyPointsRaw) > 0 {
			if err := json.Unmarshal(keyPointsRaw, &item.KeyPoints); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal key_points: %w", err)
			}
		}
		items = append(items, item)
	}

	var total int
	countQuery := "SELECT COUNT(DISTINCT fi.id)\n" + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count feed entrys: %w", err)
	}

	return items, total, nil
}

func (r *FeedRepository) GetByIDAnon(ctx context.Context, feedEntryID int) (*FeedEntryRow, error) {
	query := `
		SELECT
			fi.id AS feed_entry_id,
			fi.published_at,
			fi.title,
			fi.short_text,
			fi.key_points,
			fi.political_score,
			fi.impact_score,
			fi.source_url,
			COALESCE(agg.likes_count, 0) AS likes_count,
			COALESCE(agg.dislikes_count, 0) AS dislikes_count
		FROM feed_entries fi
		LEFT JOIN (
			SELECT
				feed_entry_id,
				SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END) AS likes_count,
				SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END) AS dislikes_count
			FROM likes
			GROUP BY feed_entry_id
		) agg ON agg.feed_entry_id = fi.id
		WHERE fi.id = $1
	`

	var item FeedEntryRow
	var keyPointsRaw []byte
	var politicalScore sql.NullInt64
	var impactScore sql.NullString
	err := r.db.QueryRowContext(ctx, query, feedEntryID).Scan(
		&item.FeedEntryID,
		&item.PublishedAt,
		&item.Title,
		&item.ShortText,
		&keyPointsRaw,
		&politicalScore,
		&impactScore,
		&item.SourceURL,
		&item.LikesCount,
		&item.DislikesCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feed entry: %w", err)
	}
	if politicalScore.Valid {
		ps := int(politicalScore.Int64)
		item.PoliticalScore = &ps
	}
	if impactScore.Valid {
		item.ImpactScore = &impactScore.String
	}
	if len(keyPointsRaw) > 0 {
		if err := json.Unmarshal(keyPointsRaw, &item.KeyPoints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal key_points: %w", err)
		}
	}
	return &item, nil
}

func (r *FeedRepository) GetByIDForUser(ctx context.Context, userID, feedEntryID int) (*FeedEntryRow, error) {
	query := `
		SELECT
			fi.id AS feed_entry_id,
			fi.published_at,
			fi.title,
			fi.short_text,
			fi.key_points,
			fi.political_score,
			fi.impact_score,
			fi.source_url,
			COALESCE(agg.likes_count, 0) AS likes_count,
			COALESCE(agg.dislikes_count, 0) AS dislikes_count,
			(CASE WHEN b.feed_entry_id IS NULL THEN FALSE ELSE TRUE END) AS is_bookmarked,
			ul.value AS user_like_status
		FROM feed_entries fi
		LEFT JOIN (
			SELECT
				feed_entry_id,
				SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END) AS likes_count,
				SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END) AS dislikes_count
			FROM likes
			GROUP BY feed_entry_id
		) agg ON agg.feed_entry_id = fi.id
		LEFT JOIN bookmarks b ON b.feed_entry_id = fi.id AND b.user_id = $2
		LEFT JOIN likes ul ON ul.feed_entry_id = fi.id AND ul.user_id = $2
		WHERE fi.id = $1
	`

	var item FeedEntryRow
	var keyPointsRaw []byte
	var politicalScore sql.NullInt64
	var impactScore sql.NullString
	var isBookmarked bool
	var userLikeStatus sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, feedEntryID, userID).Scan(
		&item.FeedEntryID,
		&item.PublishedAt,
		&item.Title,
		&item.ShortText,
		&keyPointsRaw,
		&politicalScore,
		&impactScore,
		&item.SourceURL,
		&item.LikesCount,
		&item.DislikesCount,
		&isBookmarked,
		&userLikeStatus,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feed entry for user: %w", err)
	}
	if politicalScore.Valid {
		ps := int(politicalScore.Int64)
		item.PoliticalScore = &ps
	}
	if impactScore.Valid {
		item.ImpactScore = &impactScore.String
	}
	bookmarked := isBookmarked
	item.IsBookmarked = &bookmarked
	if userLikeStatus.Valid {
		uls := int(userLikeStatus.Int64)
		item.UserLikeStatus = &uls
	}
	if len(keyPointsRaw) > 0 {
		if err := json.Unmarshal(keyPointsRaw, &item.KeyPoints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal key_points: %w", err)
		}
	}
	return &item, nil
}

func (r *FeedRepository) GetByPolicyDocID(ctx context.Context, policyDocID int) (*FeedEntryRow, error) {
	query := `
		SELECT
			fi.id AS feed_entry_id,
			fi.published_at,
			fi.title,
			fi.short_text,
			fi.key_points,
			fi.political_score,
			fi.impact_score,
			fi.source_url,
			COALESCE(agg.likes_count, 0) AS likes_count,
			COALESCE(agg.dislikes_count, 0) AS dislikes_count
		FROM feed_entries fi
		LEFT JOIN (
			SELECT
				feed_entry_id,
				SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END) AS likes_count,
				SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END) AS dislikes_count
			FROM likes
			GROUP BY feed_entry_id
		) agg ON agg.feed_entry_id = fi.id
		WHERE fi.policy_document_id = $1
	`

	var item FeedEntryRow
	var keyPointsRaw []byte
	var politicalScore sql.NullInt64
	var impactScore sql.NullString
	err := r.db.QueryRowContext(ctx, query, policyDocID).Scan(
		&item.FeedEntryID,
		&item.PublishedAt,
		&item.Title,
		&item.ShortText,
		&keyPointsRaw,
		&politicalScore,
		&impactScore,
		&item.SourceURL,
		&item.LikesCount,
		&item.DislikesCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feed entry by policy doc id: %w", err)
	}
	if politicalScore.Valid {
		ps := int(politicalScore.Int64)
		item.PoliticalScore = &ps
	}
	if impactScore.Valid {
		item.ImpactScore = &impactScore.String
	}
	if len(keyPointsRaw) > 0 {
		if err := json.Unmarshal(keyPointsRaw, &item.KeyPoints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal key_points: %w", err)
		}
	}
	return &item, nil
}

func (r *FeedRepository) UpsertFeedEntryByPolicyDocID(ctx context.Context, tx *sql.Tx, policyDocID int, title, shortText string, keyPoints []string, politicalScore *int, impactScore, sourceURL string, publishedAt time.Time) error {
	var keyPointsJSON []byte
	var err error
	if len(keyPoints) > 0 {
		keyPointsJSON, err = json.Marshal(keyPoints)
		if err != nil {
			return fmt.Errorf("failed to marshal keypoints: %w", err)
		}
	}

	var impactScorePtr *string
	if impactScore != "" {
		impactScorePtr = &impactScore
	}

	query := `
		INSERT INTO feed_entries (
			policy_document_id, title, short_text, key_points,
			political_score, impact_score, source_url, published_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (policy_document_id) DO UPDATE SET
			title           = EXCLUDED.title,
			short_text      = EXCLUDED.short_text,
			key_points      = EXCLUDED.key_points,
			political_score = EXCLUDED.political_score,
			impact_score    = EXCLUDED.impact_score,
			source_url      = EXCLUDED.source_url,
			published_at    = EXCLUDED.published_at,
			updated_at      = NOW()
	`

	_, err = tx.ExecContext(ctx, query,
		policyDocID, title, shortText, keyPointsJSON, politicalScore, impactScorePtr, sourceURL, publishedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert feed entry: %w", err)
	}
	return nil
}

func (r *FeedRepository) GetBookmarkedFeed(ctx context.Context, userID int) ([]FeedEntryRow, error) {
	query := `
		SELECT
			fi.id AS feed_entry_id,
			fi.published_at,
			fi.title,
			fi.short_text,
			fi.key_points,
			fi.political_score,
			fi.impact_score,
			fi.source_url,
			COALESCE(agg.likes_count, 0) AS likes_count,
			COALESCE(agg.dislikes_count, 0) AS dislikes_count,
			TRUE AS is_bookmarked,
			ul.value AS user_like_status
		FROM bookmarks b
		JOIN feed_entries fi ON fi.id = b.feed_entry_id
		LEFT JOIN (
			SELECT
				feed_entry_id,
				SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END) AS likes_count,
				SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END) AS dislikes_count
			FROM likes
			GROUP BY feed_entry_id
		) agg ON agg.feed_entry_id = fi.id
		LEFT JOIN likes ul ON ul.feed_entry_id = fi.id AND ul.user_id = $1
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarked feed entrys: %w", err)
	}
	defer rows.Close()

	var items []FeedEntryRow
	for rows.Next() {
		var item FeedEntryRow
		var keyPointsRaw []byte
		var politicalScore sql.NullInt64
		var impactScore sql.NullString
		var isBookmarked bool
		var userLikeStatus sql.NullInt64
		err := rows.Scan(
			&item.FeedEntryID,
			&item.PublishedAt,
			&item.Title,
			&item.ShortText,
			&keyPointsRaw,
			&politicalScore,
			&impactScore,
			&item.SourceURL,
			&item.LikesCount,
			&item.DislikesCount,
			&isBookmarked,
			&userLikeStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed entry: %w", err)
		}
		if politicalScore.Valid {
			ps := int(politicalScore.Int64)
			item.PoliticalScore = &ps
		}
		if impactScore.Valid {
			item.ImpactScore = &impactScore.String
		}
		bookmarked := isBookmarked
		item.IsBookmarked = &bookmarked
		if userLikeStatus.Valid {
			uls := int(userLikeStatus.Int64)
			item.UserLikeStatus = &uls
		}
		if len(keyPointsRaw) > 0 {
			if err := json.Unmarshal(keyPointsRaw, &item.KeyPoints); err != nil {
				return nil, fmt.Errorf("failed to unmarshal key_points: %w", err)
			}
		}
		items = append(items, item)
	}
	return items, nil
}
