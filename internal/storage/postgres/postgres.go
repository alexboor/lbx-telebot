package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type Storage struct {
	Pool *pgxpool.Pool
}

var ErrDSNEmpty = errors.New("empty dsn received")

func New(dsn string) (*Storage, error) {
	if len(dsn) == 0 {
		return nil, ErrDSNEmpty
	}
	log.Println("init postgres connection")

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MinConns = 3
	ctx := context.TODO()
	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	log.Println("connected to postgres")

	storage := &Storage{
		Pool: pool,
	}

	if err := storage.migrate(); err != nil {
		return nil, err
	}

	return storage, nil
}

// migrate is prepare db schema
// TODO: use normal migration
func (s *Storage) migrate() error {
	ctx := context.TODO()
	statements := []string{
		`create table if not exists word_count (
                user_id bigint,
                chat_id bigint,
                date date,
                val int,
                unique (user_id, chat_id, date)
            )`,
	}

	for _, stmt := range statements {
		_, err := s.Pool.Exec(ctx, stmt)
		if err != nil {
			return err
		}
	}

	return nil
}
