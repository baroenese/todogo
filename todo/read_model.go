//go:build !fake

package todo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	sq "github.com/Masterminds/squirrel"
)

type TodoList struct {
	Items []TodoItem `json:"items"`
	Count int        `json:"count"`
}

var emptyList = TodoList{}

func findAllItems(ctx context.Context, tx pgx.Tx) (TodoList, error) {
	sqlStr, _, _ := sq.Select("COUNT(id) as cnt").
		From("todolist").
		ToSql()
	var itemCount int
	if err := tx.QueryRow(ctx, sqlStr).Scan(&itemCount); err != nil {
		log.Warn().Err(err).Msg("cannot find a count in todo list")
		return emptyList, err
	}
	if itemCount == 0 {
		return emptyList, nil
	}
	log.Debug().Int("count", itemCount).Msg("found todo items")
	sql2Str, _, _ := sq.Select("id", "title", "created_at", "done_at").
		From("todolist").
		Limit(100).
		ToSql()
	rows, err := tx.Query(ctx, sql2Str)
	if err != nil {
		return emptyList, err
	}
	defer rows.Close()
	items := []TodoItem{}
	for rows.Next() {
		var item TodoItem
		if err := rows.Scan(&item.Id, &item.Title, &item.CreatedAt, &item.DoneAt); err != nil {
			log.Warn().Err(err).Msg("cannot scan an item")
			return emptyList, err
		}
		items = append(items, item)
	}
	list := TodoList{
		Items: items,
		Count: itemCount,
	}
	return list, nil
}
