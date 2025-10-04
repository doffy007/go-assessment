package comment

import (
	"context"
	"database/sql"
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

type CommentService interface {
	CreateComment(req *Comment) (*Comment, error)
	GetComment(req *Comment) error
	UpdateComment(req *Comment, updatedFields []string, tx pgx.Tx) error
	DeleteComment(id uint64) (bool, error)
	GetComments(params *apiparams.Params) (*api.Response, error)
	GetCommentsByPostID(postID uint64) (*api.Response, error)
}

type srv struct {
	db          db.DBService
	UserService user.UserService
}

var Service CommentService

func init() {
	Service = New(db.Service, user.Service)
}

func New(db db.DBService, us user.UserService) CommentService {
	return &srv{
		db:          db,
		UserService: us,
	}
}

func (s *srv) CreateComment(req *Comment) (*Comment, error) {
	if req.ID == 0 {
		id, err := uid.New()
		if err != nil {
			return nil, err
		}
		req.ID = id
	}

	if req.Status == "" {
		req.Status = StatusActive
	}

	if err := req.validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	req.CreatedAt = pgtype.Timestamptz{Time: now, Valid: true}
	req.UpdatedAt = req.CreatedAt

	if err := s.db.Commit(nil, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			context.Background(),
			`INSERT INTO comments (
				id,
				post_id,
				user_id,
				content,
				status,
				metadata,
				created_at,
				updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			req.ID,
			req.PostID,
			req.UserID,
			req.Content,
			req.Status,
			req.Metadata,
			req.CreatedAt,
			req.UpdatedAt,
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

func (s *srv) GetComment(req *Comment) error {
	if req.ID == 0 {
		return api.ErrPayload
	}

	if req.Author == nil {
		req.Author = &user.Public{}
	}

	args := []any{req.ID}
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(" WHERE comment_id = $%d ", len(args)))

	var username, name sql.NullString

	if err := s.db.QueryRow(
		`SELECT
			comment_id,
			post_id,
			user_id,
			author_username,
			author_name,
			content,
			status,
			metadata,
			created_at,
			updated_at,
			deleted_at
		FROM comment_view `+sb.String(),
		args...,
	).Scan(
		&req.ID,
		&req.PostID,
		&req.UserID,
		&username,
		&name,
		&req.Content,
		&req.Status,
		&req.Metadata,
		&req.CreatedAt,
		&req.UpdatedAt,
		&req.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return api.ErrNotFound
		}
		return err
	}

	req.Author.Username = username.String
	req.Author.Name = &name.String
	return nil
}

func (s *srv) UpdateComment(req *Comment, updatedFields []string, tx pgx.Tx) error {
	commentID := &Comment{ID: req.ID}
	if err := s.GetComment(commentID); err != nil {
		return err
	}

	if err := s.db.Commit(tx, func(tx pgx.Tx) error {
		req.UpdatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		args := []any{req.ID, req.UpdatedAt}
		var sb strings.Builder

		for i := range updatedFields {
			switch updatedFields[i] {
			case "content":
				if err := req.validate(); err != nil {
					return err
				}
				args = append(args, req.Content)
				sb.WriteString(fmt.Sprintf("content = $%d,", len(args)))
			case "status":
				args = append(args, req.Status)
				sb.WriteString(fmt.Sprintf("status = $%d,", len(args)))
			case "metadata":
				args = append(args, req.Metadata)
				sb.WriteString(fmt.Sprintf("metadata = $%d,", len(args)))
			case "deleted_at":
				args = append(args, req.DeletedAt)
				sb.WriteString(fmt.Sprintf("deleted_at = $%d,", len(args)))
			}
		}

		if len(args) > 2 {
			if _, err := tx.Exec(
				context.Background(),
				fmt.Sprintf(
					`UPDATE comments 
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

	return s.GetComment(req)
}

func (s *srv) DeleteComment(id uint64) (bool, error) {
	comment := &Comment{ID: id}
	if err := s.GetComment(comment); err != nil {
		return false, err
	}

	if err := s.db.Commit(nil, func(tx pgx.Tx) error {
		now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
		if _, err := tx.Exec(
			context.Background(),
			`UPDATE comments
			 SET updated_at = $2,
			     deleted_at = $2,
			     status = $3
			 WHERE id = $1`,
			id,
			now,
			StatusDeleted,
		); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return false, errors.New("failed to delete comment")
	}

	return true, nil
}

func (s *srv) GetComments(params *apiparams.Params) (*api.Response, error) {
	resp := &api.Response{}
	var sb strings.Builder
	args := []any{}

	if params.Search != nil && *params.Search != "" {
		sb.WriteString(" AND (content ILIKE " + db.QuoteString("%"+*params.Search+"%") + ") ")
	}

	for i := range params.Filters {
		switch {
		case strings.HasPrefix(params.Filters[i].Column, "month"):
			params.Filters[i].ColumnOverridden = true
			if params.Filters[i].Operator == apiparams.In {
				months := strings.Split(params.Filters[i].Value, ",")
				sb.WriteString(` AND EXTRACT(MONTH FROM a.created_at) IN (`)
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
				sb.WriteString(fmt.Sprintf(` AND EXTRACT(MONTH FROM a.created_at) = $%d `, len(args)))
			}

		case strings.HasPrefix(params.Filters[i].Column, "year"):
			params.Filters[i].ColumnOverridden = true
			if params.Filters[i].Operator == apiparams.In {
				years := strings.Split(params.Filters[i].Value, ",")
				sb.WriteString(` AND EXTRACT(YEAR FROM a.created_at) IN (`)
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
				sb.WriteString(fmt.Sprintf(` AND EXTRACT(YEAR FROM a.created_at) = $%d `, len(args)))
			}
		}
	}

	if len(params.DateRange) == 2 {
		start := params.DateRange[0]
		end := params.DateRange[1]
		if start.Valid && end.Valid {
			args = append(args, start.Time, end.Time)
			sb.WriteString(fmt.Sprintf(" AND a.created_at BETWEEN $%d AND $%d ", len(args)-1, len(args)))
		}
	}

	args = params.ComposeDbQueryFromFilters(&sb, args)

	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM comment_view a WHERE TRUE `+sb.String(),
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
		sb.WriteString(" ORDER BY a.created_at DESC ")
	}

	rows, err := s.db.Query(
		`SELECT
			comment_id,
			post_id,
			post_title,
			post_slug,
			user_id,
			author_username,
			author_name,
			content,
			status,
			metadata,
			created_at,
			updated_at,
			deleted_at
		FROM comment_view a
		WHERE TRUE `+sb.String()+params.Page.Compose(),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Comment
	for rows.Next() {
		var c Comment
		c.Author = &user.Public{}
		if err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.PostTitle,
			&c.PostSlug,
			&c.UserID,
			&c.Author.Username,
			&c.Author.Name,
			&c.Content,
			&c.Status,
			&c.Metadata,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			return nil, err
		}

		res = append(res, c)
	}

	resp.Results = res
	return resp, nil
}

func (s *srv) GetCommentsByPostID(postID uint64) (*api.Response, error) {
	if postID == 0 {
		return nil, errors.New("invalid post id")
	}

	resp := &api.Response{}

	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM comment_view WHERE post_id = $1`,
		postID,
	).Scan(&resp.Total); err != nil {
		return nil, err
	}

	if resp.Total == 0 {
		return resp, nil
	}

	rows, err := s.db.Query(
		`SELECT
			comment_id,
			post_id,
			user_id,
			author_username,
			author_name,
			content,
			status,
			metadata,
			created_at,
			updated_at,
			deleted_at
		FROM comment_view
		WHERE post_id = $1
		ORDER BY created_at DESC`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		c := Comment{
			Author: &user.Public{},
		}

		if err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.UserID,
			&c.Author.Username,
			&c.Author.Name,
			&c.Content,
			&c.Status,
			&c.Metadata,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	resp.Results = comments
	return resp, nil
}
