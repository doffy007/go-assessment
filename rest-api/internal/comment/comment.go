package comment

import (
	"errors"
	"rest-api/internal/user"

	"github.com/jackc/pgx/v5/pgtype"
)

type Comment struct {
	ID        uint64             `json:"id,string"`
	PostID    uint64             `json:"post_id,string"`
	PostTitle string             `json:"post_title"`
	PostSlug  string             `json:"post_slug"`
	UserID    uint64             `json:"user_id,string"`
	Author    *user.Public       `json:"author"`
	Content   *string            `json:"content"`
	Status    string             `json:"status"`
	Metadata  map[string]any     `json:"metadata"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	DeletedAt pgtype.Timestamptz `json:"deleted_at"`
}

func (c *Comment) validate() error {
	if c.Content == nil || len(*c.Content) == 0 {
		return ErrMissingContent
	}
	if len(*c.Content) > 500 {
		return ErrInvalidContent
	}
	return nil
}

const (
	StatusActive  = "active"
	StatusDeleted = "deleted"
	StatusHidden  = "hidden"
)

var (
	ErrMissingContent = errors.New("missing content")
	ErrInvalidContent = errors.New("content too long, max 500 chars")
)
