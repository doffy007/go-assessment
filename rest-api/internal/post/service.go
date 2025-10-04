package post

import (
	"context"
	"errors"
	"fmt"
	"rest-api/internal/api"
	"rest-api/internal/db"
	apiparams "rest-api/internal/params"
	"rest-api/internal/uid"
	"rest-api/internal/user"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostService interface {
	CreatePost(req *Post) (*Post, error)
	GetPost(req *Post) error
	UpdatePost(req *Post, updatedFields []string, tx pgx.Tx) error
	DeletePost(id uint64) (bool, error)
	GetPosts(params *apiparams.Params) (*api.Response, error)
}

type srv struct {
	db          db.DBService
	UserService user.UserService
}

var Service PostService

func init() {
	Service = New(db.Service, user.Service)
}

func New(db db.DBService, us user.UserService) PostService {
	return &srv{
		db:          db,
		UserService: us,
	}
}

func (s *srv) CreatePost(req *Post) (*Post, error) {
	if req.ID == 0 {
		id, err := uid.New()
		if err != nil {
			return nil, err
		}
		req.ID = id
	}

	now := time.Now()
	req.CreatedAt = pgtype.Timestamptz{Time: now, Valid: true}
	req.UpdatedAt = req.CreatedAt

	if err := s.db.Commit(nil, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			context.Background(),
			`INSERT INTO posts (
            id,
            user_id,
            type,
            title,
            slug,
            excerpt,
            seo_title,
            seo_excerpt,
            content,
            content_html,
            content_text,
            tags,
            status,
            created_at,
            updated_at,
            scheduled_at,
            published_at,
            total_views,
            metadata
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
			req.ID,
			req.UserID,
			req.Type,
			req.Title,
			req.Slug,
			req.Excerpt,
			req.SEOTitle,
			req.SEOExcerpt,
			req.Content,
			req.ContentHTML,
			req.ContentText,
			req.Tags,
			req.Status,
			req.CreatedAt,
			req.UpdatedAt,
			req.ScheduledAt,
			req.PublishedAt,
			req.TotalViews,
			req.Metadata,
		)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *srv) GetPost(req *Post) error {
	if req.ID == 0 {
		return api.ErrPayload
	}

	args := []any{}
	sb := strings.Builder{}

	args = append(args, req.ID)
	sb.WriteString(fmt.Sprintf(" WHERE id = $%d ", len(args)))

	if err := s.db.QueryRow(
		`SELECT
			id,
			user_id,
			type,
			title,
			slug,
			excerpt,
			seo_title,
			seo_excerpt,
			content,
			content_html,
			content_text,
			tags,
			status,
			created_at,
			updated_at,
			scheduled_at,
			published_at,
			total_views,
			metadata
		FROM posts `+sb.String(),
		args...,
	).Scan(
		&req.ID,
		&req.UserID,
		&req.Type,
		&req.Title,
		&req.Slug,
		&req.Excerpt,
		&req.SEOTitle,
		&req.SEOExcerpt,
		&req.Content,
		&req.ContentHTML,
		&req.ContentText,
		&req.Tags,
		&req.Status,
		&req.CreatedAt,
		&req.UpdatedAt,
		&req.ScheduledAt,
		&req.PublishedAt,
		&req.TotalViews,
		&req.Metadata,
	); err != nil {
		if err == pgx.ErrNoRows {
			return api.ErrNotFound
		}
		return err
	}

	if req.UserID != 0 {
		author, err := s.UserService.GetUser(req.UserID)
		if err != nil {
			return err
		}
		req.Author = user.ToPublic(author)
	}

	return nil
}

func (s *srv) UpdatePost(req *Post, updatedFields []string, tx pgx.Tx) error {
	postID := &Post{ID: req.ID}
	if err := s.GetPost(postID); err != nil {
		return err
	}

	if err := s.db.Commit(tx, func(tx pgx.Tx) error {
		req.UpdatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		args := []any{req.ID, req.UpdatedAt}
		var sb strings.Builder

		for i := range updatedFields {
			switch updatedFields[i] {
			case "type":
				args = append(args, req.Type)
				sb.WriteString(fmt.Sprintf("type = $%d,", len(args)))
			case "title":
				args = append(args, req.Title)
				sb.WriteString(fmt.Sprintf("title = $%d,", len(args)))
			case "slug":
				args = append(args, req.Slug)
				sb.WriteString(fmt.Sprintf("slug = $%d,", len(args)))
			case "excerpt":
				args = append(args, req.Excerpt)
				sb.WriteString(fmt.Sprintf("excerpt = $%d,", len(args)))
			case "seo_title":
				args = append(args, req.SEOTitle)
				sb.WriteString(fmt.Sprintf("seo_title = $%d,", len(args)))
			case "seo_excerpt":
				args = append(args, req.SEOExcerpt)
				sb.WriteString(fmt.Sprintf("seo_excerpt = $%d,", len(args)))
			case "content":
				args = append(args, req.Content)
				sb.WriteString(fmt.Sprintf("content = $%d,", len(args)))
			case "content_html":
				args = append(args, req.ContentHTML)
				sb.WriteString(fmt.Sprintf("content_html = $%d,", len(args)))
			case "content_text":
				args = append(args, req.ContentText)
				sb.WriteString(fmt.Sprintf("content_text = $%d,", len(args)))
			case "tags":
				args = append(args, req.Tags)
				sb.WriteString(fmt.Sprintf("tags = $%d,", len(args)))
			case "status":
				args = append(args, req.Status)
				sb.WriteString(fmt.Sprintf("status = $%d,", len(args)))
			case "scheduled_at":
				args = append(args, req.ScheduledAt)
				sb.WriteString(fmt.Sprintf("scheduled_at = $%d,", len(args)))
			case "published_at":
				args = append(args, req.PublishedAt)
				sb.WriteString(fmt.Sprintf("published_at = $%d,", len(args)))
			case "metadata":
				args = append(args, req.Metadata)
				sb.WriteString(fmt.Sprintf("metadata = $%d,", len(args)))
			}
		}

		if len(args) > 2 {
			if _, err := tx.Exec(
				context.Background(),
				fmt.Sprintf(
					`UPDATE posts 
						SET %s updated_at = $2 
					WHERE id = $1`, sb.String()),
				args...,
			); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return s.GetPost(req)
}

func (s *srv) DeletePost(id uint64) (bool, error) {
	postID := &Post{ID: id}
	if err := s.GetPost(postID); err != nil {
		return false, err
	}

	if err := s.db.Commit(nil, func(tx pgx.Tx) error {
		now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
		if _, err := tx.Exec(
			context.Background(),
			`UPDATE posts
			 SET updated_at = $2,
			     deleted_at = $2
			 WHERE id = $1`,
			id,
			now,
		); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return false, errors.New("failed to delete post")
	}

	return true, nil
}

func (s *srv) GetPosts(params *apiparams.Params) (*api.Response, error) {
	resp := &api.Response{}
	var sb strings.Builder
	args := []any{}

	if params.Search != nil && *params.Search != "" {
		sb.WriteString(" AND (title || ' ') ILIKE " + db.QuoteString("%"+*params.Search+"%") + " ")
	}

	for i := range params.Filters {
		switch {
		case strings.HasPrefix(params.Filters[i].Column, "month"):
			params.Filters[i].ColumnOverridden = true
			if params.Filters[i].Operator == apiparams.In {
				months := strings.Split(params.Filters[i].Value, ",")
				sb.WriteString(` AND EXTRACT(MONTH FROM created_at) IN (`)
				for j, month := range months {
					monthValue, err := time.Parse("January", strings.Title(strings.TrimSpace(month)))
					if err != nil {
						return nil, errors.New("invalid month format: " + month)
					}
					args = append(args, monthValue.Month())
					if j > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("$%d", len(args)))
				}
				sb.WriteString(") ")
			} else {
				monthValue, err := time.Parse("January", strings.Title(strings.TrimSpace(params.Filters[i].Value)))
				if err != nil {
					return nil, errors.New("invalid month format: " + params.Filters[i].Value)
				}
				args = append(args, monthValue.Month())
				sb.WriteString(fmt.Sprintf(` AND EXTRACT(MONTH FROM created_at) = $%d `, len(args)))
			}

		case strings.HasPrefix(params.Filters[i].Column, "year"):
			params.Filters[i].ColumnOverridden = true
			if params.Filters[i].Operator == apiparams.In {
				years := strings.Split(params.Filters[i].Value, ",")
				sb.WriteString(` AND EXTRACT(YEAR FROM created_at) IN (`)
				for j, year := range years {
					args = append(args, strings.TrimSpace(year))
					if j > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("$%d", len(args)))
				}
				sb.WriteString(") ")
			} else {
				args = append(args, strings.TrimSpace(params.Filters[i].Value))
				sb.WriteString(fmt.Sprintf(` AND EXTRACT(YEAR FROM created_at) = $%d `, len(args)))
			}
		}
	}

	if len(params.DateRange) == 2 {
		start := params.DateRange[0]
		end := params.DateRange[1]
		if start.Valid && end.Valid {
			args = append(args, start.Time, end.Time)
			sb.WriteString(fmt.Sprintf(" AND created_at BETWEEN $%d AND $%d ", len(args)-1, len(args)))
		}
	}

	args = params.ComposeDbQueryFromFilters(&sb, args)

	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL `+sb.String(),
		args...,
	).Scan(&resp.Total); err != nil {
		return nil, err
	}

	if resp.Total == 0 {
		return resp, nil
	}

	if len(params.Sorts) > 0 {
		sb.WriteString(" ORDER BY ")
		for i, o := range params.Sorts {
			sb.WriteString(fmt.Sprintf(" %s %s ", o.Column, db.Order(o.Asc)))
			if i < len(params.Sorts)-1 {
				sb.WriteString(",")
			}
		}
	} else {
		sb.WriteString(" ORDER BY created_at DESC ")
	}

	rows, err := s.db.Query(
		`SELECT
			id,
			user_id,
			type,
			title,
			slug,
			excerpt,
			seo_title,
			seo_excerpt,
			content,
			content_html,
			content_text,
			tags,
			status,
			created_at,
			updated_at,
			scheduled_at,
			published_at,
			total_views,
			metadata
		FROM posts
		WHERE TRUE `+sb.String()+params.Page.Compose(),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Type,
			&p.Title,
			&p.Slug,
			&p.Excerpt,
			&p.SEOTitle,
			&p.SEOExcerpt,
			&p.Content,
			&p.ContentHTML,
			&p.ContentText,
			&p.Tags,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.ScheduledAt,
			&p.PublishedAt,
			&p.TotalViews,
			&p.Metadata,
		); err != nil {
			return nil, err
		}

		if p.UserID != 0 {
			author, err := user.Service.GetUser(p.UserID)
			if err == nil {
				p.Author = user.ToPublic(author)
			}
		}

		res = append(res, p)
	}

	resp.Results = res
	return resp, nil
}
