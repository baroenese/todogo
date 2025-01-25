//go:build !fake

package todo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"

	sq "github.com/Masterminds/squirrel"
)

func findItemById(ctx context.Context, tx pgx.Tx, id ulid.ULID) (TodoItem, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sqlStr, args, err := psql.
		Select("id", "title", "created_at", "done_at").
		From("todolist").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		log.Fatal().Err(err).Msg("sql error")
		return TodoItem{}, err
	}
	var item TodoItem
	err = tx.QueryRow(ctx, sqlStr, args...).Scan(&item.Id, &item.Title, &item.CreatedAt, &item.DoneAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Fatal().Err(err).Msg("can't find any item")
			return TodoItem{}, err
		}
		return TodoItem{}, err
	}
	return item, nil
}

func saveItem(ctx context.Context, tx pgx.Tx, item TodoItem) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sqlStrCreate, args, err := psql.
		Insert("todolist").
		Columns("id", "title", "created_at", "done_at").
		Values(item.Id, item.Title, item.CreatedAt, item.DoneAt).
		Suffix("ON CONFLICT(id) DO UPDATE SET title = EXCLUDED.title, done_at = EXCLUDED.done_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Fatal().Err(err).Msg("sql error")
		return err
	}
	_, execErr := tx.Exec(ctx, sqlStrCreate, args...)
	if execErr != nil {
		return execErr
	}
	return nil
}
