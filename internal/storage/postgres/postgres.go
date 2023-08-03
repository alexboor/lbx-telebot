package postgres

import (
	"context"
	"errors"
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
	// TODO: add primary key for user_id
	queryCreateWordCount := `
create table if not exists word_count
(
    user_id bigint,
    chat_id bigint,
    date    date,
    val     int,
    unique (user_id, chat_id, date)
)`

	queryCreateProfile := `
create table if not exists profile
(
    id         numeric default 0  not null constraint profile_pk primary key,
    first_name text    default '' not null,
    last_name  text    default '' not null,
    user_name   text    default '' not null
)`

	statements := []string{queryCreateWordCount, queryCreateProfile}
	for _, stmt := range statements {
		if _, err := s.Pool.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	return nil
}
