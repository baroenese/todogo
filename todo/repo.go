//go:build !fake

package todo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"

	sq "github.com/Masterminds/squirrel"
)

func findItemById(ctx context.Context, tx pgx.Tx, id ulid.ULID) (TodoItem, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sqlStr, _, _ := psql.Select("id", "title", "created_at", "done_at").
		From("todolist").
		Where("id = ?", id).
		ToSql()
	row := tx.QueryRow(ctx, sqlStr, id)
	var item TodoItem
	if err := row.Scan(&item.Id, &item.Title, &item.CreatedAt, &item.DoneAt); err != nil {
		if err == pgx.ErrNoRows {
			log.Debug().Err(err).Msg("can't find any item")
			return TodoItem{}, err
		}
		return TodoItem{}, err
	}
	return item, nil
}

func saveItem(ctx context.Context, tx pgx.Tx, item TodoItem) error {
	sqlStrCreate, args, _ := sq.Insert("todolist").
		Columns("id", "title", "created_at", "done_at").
		Values(item.Id, item.Title, item.CreatedAt, item.DoneAt).
		Suffix("ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title, done_at=EXCLUDED.done_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	_, err := tx.Exec(ctx, sqlStrCreate, args...)
	if err != nil {
		return err
	}
	return nil
}
