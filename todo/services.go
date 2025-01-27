package todo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

func withTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func listItems(ctx context.Context) (TodoList, error) {
	var list TodoList
	err := withTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		list, err = findAllItems(ctx, tx)
		return err
	})
	return list, err
}

func createItem(ctx context.Context, title string) (ulid.ULID, error) {
	var id ulid.ULID
	todoItem, err := NewTodoItem(title)
	if err != nil {
		return id, err
	}
	err = withTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return saveItem(ctx, tx, todoItem)
	})
	if err != nil {
		return id, err
	}
	return todoItem.Id, nil
}

func findItem(ctx context.Context, id ulid.ULID) (item TodoItem, err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return
	}
	item, err = findItemById(ctx, tx, id)
	if err != nil {
		return TodoItem{}, err
	}
	err = tx.Commit(ctx)
	return
}

func makeItemDone(ctx context.Context, id ulid.ULID) error {
	return withTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		item, err := findItemById(ctx, tx, id)
		if err != nil {
			return err
		}
		if err = item.MakeDone(); err != nil {
			tx.Rollback(ctx)
			return err
		}
		return saveItem(ctx, tx, item)
	})
}
