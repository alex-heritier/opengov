package services

import (
	"context"

	"github.com/alex/opengov-go/internal/repository"
	"github.com/alex/opengov-go/internal/timeformat"
	"github.com/alex/opengov-go/internal/transport"
)

type FeedService struct {
	feedRepo *repository.FeedRepository
}

func NewFeedService(feedRepo *repository.FeedRepository) *FeedService {
	return &FeedService{feedRepo: feedRepo}
}

func (s *FeedService) GetFeed(ctx context.Context, userID *int, page, limit int, sort string) (transport.FeedResponse, error) {
	var items []repository.FeedEntryRow
	var total int
	var err error

	if userID != nil {
		items, total, err = s.feedRepo.GetFeedForUser(ctx, *userID, page, limit, sort)
	} else {
		items, total, err = s.feedRepo.GetFeedAnon(ctx, page, limit, sort)
	}

	if err != nil {
		return transport.FeedResponse{}, err
	}

	responses := make([]transport.FeedEntryResponse, len(items))
	for i, item := range items {
		responses[i] = mapFeedEntryRowToResponse(item)
	}

	offset := (page - 1) * limit
	return transport.FeedResponse{
		Items:   responses,
		Page:    page,
		Limit:   limit,
		Total:   total,
		HasNext: offset+limit < total,
	}, nil
}

func (s *FeedService) GetItem(ctx context.Context, userID *int, feedEntryID int) (*transport.FeedEntryResponse, error) {
	var item *repository.FeedEntryRow
	var err error

	if userID != nil {
		item, err = s.feedRepo.GetByIDForUser(ctx, *userID, feedEntryID)
	} else {
		item, err = s.feedRepo.GetByIDAnon(ctx, feedEntryID)
	}

	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	resp := mapFeedEntryRowToResponse(*item)
	return &resp, nil
}

func (s *FeedService) GetBookmarkedFeed(ctx context.Context, userID int) ([]transport.FeedEntryResponse, error) {
	items, err := s.feedRepo.GetBookmarkedFeed(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]transport.FeedEntryResponse, len(items))
	for i, item := range items {
		responses[i] = mapFeedEntryRowToResponse(item)
	}
	return responses, nil
}

func mapFeedEntryRowToResponse(item repository.FeedEntryRow) transport.FeedEntryResponse {
	return transport.FeedEntryResponse{
		ID:             item.FeedEntryID,
		Title:          item.Title,
		Summary:        item.ShortText,
		Keypoints:      item.KeyPoints,
		ImpactScore:    item.ImpactScore,
		PoliticalScore: item.PoliticalScore,
		SourceURL:      item.SourceURL,
		PublishedAt:    item.PublishedAt.Format(timeformat.DBTime),
		IsBookmarked:   item.IsBookmarked,
		UserLikeStatus: item.UserLikeStatus,
		LikesCount:     item.LikesCount,
		DislikesCount:  item.DislikesCount,
	}
}
