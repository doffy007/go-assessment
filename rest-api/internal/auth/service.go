package auth

import (
	"context"
	"rest-api/internal/db"
	"rest-api/internal/jwt"
	"rest-api/internal/uid"
	"rest-api/internal/user"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthService interface {
	Register(req *user.User) (*user.User, error)
	SignIn(username, password string) (string, error)
}

type srv struct {
	db db.DBService
}

var Service AuthService

func init() {
	Service = New(db.Service)
}

func New(db db.DBService) AuthService {
	return &srv{db}
}

func (s *srv) Register(req *user.User) (*user.User, error) {
	if req.ID == 0 {
		id, err := uid.New()
		if err != nil {
			return nil, err
		}
		req.ID = id
	}

	ok, err := user.Service.IsUsernameOrEmailExists(req.Username, req.Email)
	if ok {
		return nil, err
	}

	if err := req.ValidateEmail(); err != nil {
		return nil, err
	}

	hashed, err := user.HashPassword(*req.Password)
	if err != nil {
		return nil, err
	}
	*req.Password = hashed

	req.CreatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	req.UpdatedAt = req.CreatedAt

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		context.Background(),
		`INSERT INTO users (
			id,
			name,
			username,
			password,
			bio,
			email,
			language,
			created_at,
			updated_at,
			email_verified,
			country_code,
			phone_number,
			location,
			nationality,
			national_id_number
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		req.ID,
		req.Name,
		req.Username,
		req.Password,
		req.Bio,
		req.Email,
		req.Language,
		req.CreatedAt,
		req.UpdatedAt,
		req.EmailVerified,
		req.CountryCode,
		req.PhoneNumber,
		req.Location,
		req.Nationality,
		req.NationalIDNumber,
	)
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}

	req.RemoveSensitivePII()

	return req, nil
}

func (s *srv) SignIn(username, password string) (string, error) {
	var result user.User

	err := s.db.QueryRow(
		`SELECT
			id,
			username,
			password,
			email,
			created_at,
			updated_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL`,
		username,
	).Scan(
		&result.ID,
		&result.Username,
		&result.Password,
		&result.Email,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", user.ErrUserNotFound
		}
		return "", err
	}

	// check password
	err = user.CompareHashAndPassword(*result.Password, password)
	if err != nil {
		return "", err
	}

	token, err := jwt.GenerateJWT(result.Username, result.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}
