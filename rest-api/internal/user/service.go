package user

import (
	"context"
	"errors"
	"fmt"
	"rest-api/internal/api"
	"rest-api/internal/db"
	"rest-api/internal/jwt"
	"rest-api/internal/uid"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	gonanoid "github.com/matoous/go-nanoid/v2"

	apiparams "rest-api/internal/params"
)

type UserService interface {
	CreateUser(req *User) (*User, error)
	IsUsernameOrEmailExists(username string, email string) (bool, error)
	GetUser(id uint64) (*User, error)
	GetUsers(params *apiparams.Params) (*api.Response, error)
	UpdateUser(req *User, updatedFields []string) (*User, error)
	DeleteUser(id uint64) (bool, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByToken(token string) (*User, error)
}

type srv struct {
	db db.DBService
}

var Service UserService

func init() {
	Service = New(db.Service)
}

func New(db db.DBService) UserService {
	return &srv{db}
}

func (s *srv) CreateUser(req *User) (*User, error) {
	if req.ID == 0 {
		id, err := uid.New()
		if err != nil {
			return nil, err
		}
		req.ID = id
	}

	ok, err := s.IsUsernameOrEmailExists(req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, ErrUsernameOrEmailExists
	}

	if err := req.ValidateEmail(); err != nil {
		return nil, err
	}

	now := time.Now()
	req.CreatedAt = pgtype.Timestamptz{Time: now, Valid: true}
	req.UpdatedAt = req.CreatedAt

	rawPassword, err := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 12)
	if err != nil {
		return nil, err
	}

	hashpass, err := HashPassword(rawPassword)
	if err != nil {
		return nil, err
	}

	req.Password = &hashpass

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

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
		return nil, err
	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}

	req.RemoveSensitivePII()
	return req, nil
}

func (s *srv) IsUsernameOrEmailExists(username, email string) (bool, error) {
	var exists bool

	if username != "" {
		if err := s.db.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)`,
			username,
		).Scan(&exists); err != nil {
			return false, errors.New("failed to check username existence")
		}

		if exists {
			return true, errors.New("username already exists")
		}
	}

	if email != "" {
		if err := s.db.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`,
			email,
		).Scan(&exists); err != nil {
			return false, errors.New("failed to check email existence")
		}

		if exists {
			return true, errors.New("email already exists")
		}
	}

	return false, nil
}

func (s *srv) GetUsers(params *apiparams.Params) (*api.Response, error) {
	resp := &api.Response{}

	var sb strings.Builder
	if params.Search != nil && *params.Search != "" {
		sb.WriteString(" AND (COALESCE(name, '') || ' ' || COALESCE(username, '') || ' ' || email) ILIKE " + db.QuoteString("%"+*params.Search+"%") + " ")
	}

	args := params.ComposeDbQueryFromFilters(&sb, []any{})

	if params.Days != nil {
		sb.WriteString(fmt.Sprintf(" AND created_at::DATE > CURRENT_DATE - INTERVAL '%d day' ", *params.Days))
	} else if params.DateRange != nil {
		args = append(args, params.DateRange[0], params.DateRange[1])
		n := len(args)
		sb.WriteString(fmt.Sprintf(" AND created_at::DATE BETWEEN $%d AND $%d ", n-1, n))
	}

	if err := s.db.QueryRow(
		`SELECT COUNT(*) 
		FROM users 
		WHERE deleted_at IS NULL `+sb.String(),
		args...,
	).Scan(&resp.Total); err != nil {
		return nil, err
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
			name,
			username,
			bio,
			email,
			email_verified,
			role,
			gender::VARCHAR,
			language,
			created_at,
			updated_at,
			deactivated_at,
			deleted_at,
			last_seen,
			ip_address,
			user_agent,
			country_code,
			phone_number,
			location,
			nationality,
			national_id_number
		FROM users
		WHERE deleted_at IS NULL `+sb.String()+params.Page.Compose(),
		args...,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return resp, nil
		}
		return nil, err
	}

	defer rows.Close()

	res := []User{}
	for rows.Next() {
		var o User

		if err := rows.Scan(
			&o.ID,
			&o.Name,
			&o.Username,
			&o.Bio,
			&o.Email,
			&o.EmailVerified,
			&o.Role,
			&o.Gender,
			&o.Language,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.DeactivatedAt,
			&o.DeletedAt,
			&o.LastSeen,
			&o.IPAddress,
			&o.UserAgent,
			&o.CountryCode,
			&o.PhoneNumber,
			&o.Location,
			&o.Nationality,
			&o.NationalIDNumber,
		); err != nil {
			return nil, err
		}

		res = append(res, o)
	}

	if len(res) == 0 {
		return resp, nil
	}

	resp.Results = res
	return resp, nil
}

func (s *srv) GetUser(id uint64) (*User, error) {
	var result User
	if err := s.db.QueryRow(
		`SELECT
			id,
			name,
			username,
			bio,
			email,
			email_verified,
			role,
			gender::VARCHAR,
			language,
			created_at,
			updated_at,
			deactivated_at,
			deleted_at,
			last_seen,
			ip_address,
			user_agent,
			country_code,
			phone_number,
			location,
			nationality,
			national_id_number
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Username,
		&result.Bio,
		&result.Email,
		&result.EmailVerified,
		&result.Role,
		&result.Gender,
		&result.Language,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeactivatedAt,
		&result.DeletedAt,
		&result.LastSeen,
		&result.IPAddress,
		&result.UserAgent,
		&result.CountryCode,
		&result.PhoneNumber,
		&result.Location,
		&result.Nationality,
		&result.NationalIDNumber,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	result.RemoveSensitivePII()
	return &result, nil
}

func (s *srv) UpdateUser(req *User, updatedFields []string) (*User, error) {
	if req.ID == 0 {
		return nil, api.ErrNotFound
	}

	req.UpdatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}

	var sb strings.Builder
	args := []any{req.ID, req.UpdatedAt}

	for i := range updatedFields {
		switch updatedFields[i] {
		case "name":
			args = append(args, req.Name)
			sb.WriteString(fmt.Sprintf("name = $%d,", len(args)))
		case "username":
			ok, _ := s.IsUsernameOrEmailExists(req.Username, req.Email)
			if ok {
				return nil, fmt.Errorf("username already exists")
			}
			args = append(args, req.Username)
			sb.WriteString(fmt.Sprintf("username = $%d,", len(args)))
		case "bio":
			args = append(args, req.Bio)
			sb.WriteString(fmt.Sprintf("bio = $%d,", len(args)))
		case "language":
			args = append(args, req.Language)
			sb.WriteString(fmt.Sprintf("language = $%d,", len(args)))
		case "gender":
			args = append(args, req.Gender)
			sb.WriteString(fmt.Sprintf("gender = $%d,", len(args)))
		case "birthdate":
			args = append(args, req.Birthdate)
			sb.WriteString(fmt.Sprintf("birthdate = $%d,", len(args)))
		case "country_code":
			args = append(args, req.CountryCode)
			sb.WriteString(fmt.Sprintf("country_code = $%d,", len(args)))
		case "phone_number":
			args = append(args, req.PhoneNumber)
			sb.WriteString(fmt.Sprintf("phone_number = $%d,", len(args)))
		case "location":
			args = append(args, req.Location)
			sb.WriteString(fmt.Sprintf("location = $%d,", len(args)))
		case "nationality":
			args = append(args, req.Nationality)
			sb.WriteString(fmt.Sprintf("nationality = $%d,", len(args)))
		case "national_id_number":
			args = append(args, req.NationalIDNumber)
			sb.WriteString(fmt.Sprintf("national_id_number = $%d,", len(args)))
		case "password":
			if req.Password == nil {
				return nil, errors.New("password cannot be nil")
			}

			if !IsValidPassword(*req.Password) {
				return nil, errors.New("invalid password format")
			}

			hashedPassword, err := HashPassword(*req.Password)
			if err != nil {
				return nil, err
			}

			req.Password = &hashedPassword

			args = append(args, req.Password)
			sb.WriteString(fmt.Sprintf("password = $%d,", len(args)))
		}
	}

	if len(args) > 0 {
		if err := s.db.Commit(nil, func(tx pgx.Tx) error {
			_, err := tx.Exec(
				context.Background(),
				fmt.Sprintf(
					`UPDATE users
					SET %s
						updated_at = $2
					WHERE id = $1`,
					sb.String(),
				),
				args...,
			)
			if err != nil {
				return err
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	res, err := s.GetUser(req.ID)
	if err != nil {
		res = req
	}

	return res, nil
}

func (s *srv) DeleteUser(id uint64) (bool, error) {
	if err := s.db.Commit(nil, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			context.Background(),
			`UPDATE users
			SET deleted_at = $2,
				updated_at = $2
			WHERE id = $1`,
			id,
			time.Now(),
		)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (s *srv) GetUserByEmail(email string) (*User, error) {
	var result User
	if err := s.db.QueryRow(
		`SELECT
			id,
			name,
			username,
			bio,
			email,
			email_verified,
			role,
			gender::VARCHAR,
			language,
			created_at,
			updated_at,
			deactivated_at,
			deleted_at,
			last_seen,
			ip_address,
			user_agent,
			country_code,
			phone_number,
			location,
			nationality,
			national_id_number
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`,
		email,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Username,
		&result.Bio,
		&result.Email,
		&result.EmailVerified,
		&result.Role,
		&result.Gender,
		&result.Language,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeactivatedAt,
		&result.DeletedAt,
		&result.LastSeen,
		&result.IPAddress,
		&result.UserAgent,
		&result.CountryCode,
		&result.PhoneNumber,
		&result.Location,
		&result.Nationality,
		&result.NationalIDNumber,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, ErrUserNotFound
	}

	return &result, nil
}

func (s *srv) GetUserByToken(token string) (*User, error) {
	claims, err := jwt.ParseToken(token)
	if err != nil {
		return nil, err
	}

	if claims.UserID == 0 {
		return nil, errors.New("invalid token claims")
	}

	uid := claims.UserID

	var result User
	if err := s.db.QueryRow(
		`SELECT
			id,
			name,
			username,
			bio,
			email,
			email_verified,
			role,
			gender::VARCHAR,
			language,
			created_at,
			updated_at,
			deactivated_at,
			deleted_at,
			last_seen,
			ip_address,
			user_agent,
			country_code,
			phone_number,
			location,
			nationality,
			national_id_number
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`,
		uid,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Username,
		&result.Bio,
		&result.Email,
		&result.EmailVerified,
		&result.Role,
		&result.Gender,
		&result.Language,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeactivatedAt,
		&result.DeletedAt,
		&result.LastSeen,
		&result.IPAddress,
		&result.UserAgent,
		&result.CountryCode,
		&result.PhoneNumber,
		&result.Location,
		&result.Nationality,
		&result.NationalIDNumber,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}
