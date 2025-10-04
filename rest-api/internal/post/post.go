package post

import (
	"errors"
	"rest-api/internal/user"

	"github.com/jackc/pgx/v5/pgtype"
)

type Post struct {
	ID          uint64             `json:"id,string"`
	Author      *user.Public       `json:"author"`
	Type        *string            `json:"type"`
	Title       *string            `json:"title"`
	Slug        *string            `json:"slug"`
	Excerpt     *string            `json:"excerpt"`
	SEOTitle    *string            `json:"seo_title"`
	SEOExcerpt  *string            `json:"seo_excerpt"`
	Content     map[string]any     `json:"content"`
	ContentHTML *string            `json:"content_html"`
	ContentText *string            `json:"content_text"`
	Tags        []string           `json:"tags"`
	Status      string             `json:"status"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
	ScheduledAt pgtype.Timestamptz `json:"scheduled_at"`
	PublishedAt pgtype.Timestamptz `json:"published_at"`
	TotalViews  int                `json:"total_views,omitempty"`
	UserID      uint64             `json:"-"`
	Metadata    map[string]any     `json:"metadata"`
}

func (a *Post) validate() error {
	if a.Slug == nil {
		return ErrMissingSlug
	}

	if a.Excerpt != nil && len(*a.Excerpt) > 150 {
		return ErrInvalidExcerpt
	}

	return nil
}

type Tag struct {
	Tag   string `json:"tag"`
	Total int64  `json:"total"`
}

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusApproved  = "approved"
	StatusPending   = "pending"
)

var (
	ErrInvalidExcerpt = errors.New("excerpt more than 150 character")
	ErrMissingSlug    = errors.New("missing slug")
	ErrStatusExists   = errors.New("exists")
)
