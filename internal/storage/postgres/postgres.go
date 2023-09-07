package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/alexboor/lbx-telebot/internal/storage/postgres/migration"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Storage struct {
	Pool *pgxpool.Pool
}

var ErrDSNEmpty = errors.New("empty dsn received")

func New(ctx context.Context, dsn string) (*Storage, error) {
	if len(dsn) == 0 {
		return nil, ErrDSNEmpty
	}
	log.Println("init postgres connection")

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MinConns = 3
	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	log.Println("connected to postgres")

	storage := &Storage{Pool: pool}

	return storage, storage.migrate(ctx)
}

// migrate is prepare db schema
// TODO: use normal migration
func (s *Storage) migrate(ctx context.Context) error {
	for _, stmt := range migration.Migrations {
		if _, err := s.Pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to make migration: %v", err)
		}
	}

	return nil
}
