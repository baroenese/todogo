//go:build !fake

package todo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"gopkg.in/guregu/null.v4"

	sq "github.com/Masterminds/squirrel"
)

type TodoList struct {
	Items []TodoItem `json:"items"`
	Count int        `json:"count"`
}

var (
	emptyList = TodoList{}
	psql2     = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

const limit = 100

func findAllItems(ctx context.Context, tx pgx.Tx) (TodoList, error) {
	sqlStr, args, err := psql2.Select("COUNT(id) as cnt").
		From("todolist").
		ToSql()
	if err != nil {
		log.Fatal().Err(err).Msg("sql1 error")
		return emptyList, err
	}
	var itemCount int
	if err := tx.QueryRow(ctx, sqlStr, args...).Scan(&itemCount); err != nil {
		log.Fatal().Err(err).Msg("cannot find a count in todo list")
		return emptyList, err
	}
	if itemCount == 0 {
		log.Debug().Msg("No todo items found")
		return emptyList, nil
	}
	log.Debug().Int("count", itemCount).Msg("found todo items")
	sql2Str, args2, err := psql.Select("id", "title", "created_at", "done_at").
		From("todolist").
		Limit(limit).
		ToSql()
	if err != nil {
		log.Error().Err(err).Msg("sql2 error")
		return emptyList, err
	}
	rows, err := tx.Query(ctx, sql2Str, args2...)
	if err != nil {
		log.Error().Err(err).Msg("failed execute")
		return emptyList, err
	}
	defer rows.Close()
	items := make([]TodoItem, 0, limit)
	for rows.Next() {
		var id ulid.ULID
		var title string
		var createdAt time.Time
		var doneAt null.Time
		if err := rows.Scan(&id, &title, &createdAt, &doneAt); err != nil {
			log.Error().Err(err).Msg("cannot scan an item")
			return emptyList, err
		}
		items = append(items, TodoItem{
			id:        id,
			title:     title,
			createdAt: createdAt,
			doneAt:    doneAt,
		})
	}
	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("Error occurred while iterating rows")
		return emptyList, err
	}
	return TodoList{
		Items: items,
		Count: itemCount,
	}, nil
}
