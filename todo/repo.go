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

var (
	psql            = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	ErrItemNotFound = errors.New("item not found")
)

func findItemById(ctx context.Context, tx pgx.Tx, idx ulid.ULID) (TodoItem, error) {
	sqlStr, args, err := psql.Select("id", "title", "created_at", "done_at").
		From("todolist").
		Where(sq.Eq{"id": idx}).
		ToSql()
	if err != nil {
		log.Error().Err(err).Msg("sql error")
		return TodoItem{}, err
	}
	var item TodoItem
	err = tx.QueryRow(ctx, sqlStr, args...).
		Scan(&item.id, &item.title, &item.createdAt, &item.doneAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error().Err(err).Msg("can't find any item")
			return TodoItem{}, ErrItemNotFound
		}
		log.Error().Err(err).Str("id", idx.String()).Msg("failed to query item")
		return TodoItem{}, err
	}
	return item, nil
}

func saveItem(ctx context.Context, tx pgx.Tx, item TodoItem) error {
	sqlStrCreate, args, err := psql.Insert("todolist").
		Columns("id", "title", "created_at", "done_at").
		Values(item.GetID(), item.GetTitle(), item.GetCreatedAt(), item.GetDoneAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET title = EXCLUDED.title, done_at = EXCLUDED.done_at").
		ToSql()
	if err != nil {
		log.Error().Err(err).
			Interface("item", item).
			Msg("sql error")
		return err
	}
	_, err = tx.Exec(ctx, sqlStrCreate, args...)
	if err != nil {
		log.Error().
			Err(err).
			Interface("item", item).
			Msg("failed to save item")
		return err
	}
	return nil
}
