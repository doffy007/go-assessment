package role

import (
	"context"
	"rest-api/internal/db"
	"time"

	"github.com/jackc/pgx/v5"
)

type RoleService interface {
	AssignRole(userId uint64, role string) error
	UnassignRole(userId uint64) error
}

type srv struct {
	db db.DBService
}

var Service RoleService

func init() {
	Service = New(db.Service)
}

func New(db db.DBService) RoleService {
	return &srv{db}
}

func (s *srv) AssignRole(userId uint64, role string) error {
	if err := s.db.Commit(nil, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			context.Background(),
			`UPDATE users
			SET role = $2,
				updated_at = $3
			WHERE id = $1`,
			userId,
			role,
			time.Now(),
		)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *srv) UnassignRole(userId uint64) error {
	if err := s.db.Commit(nil, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			context.Background(),
			`UPDATE users
			SET role = NULL,
				updated_at = $2
			WHERE id = $1`,
			userId,
			time.Now(),
		)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
