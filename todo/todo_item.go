package todo

import (
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"gopkg.in/guregu/null.v4"
)

var (
	ErrIsDone       = errors.New("todo: the item is done")
	ErrInvalidTitle = errors.New("todo: title is invalid")
)

type TodoItem struct {
	id        ulid.ULID
	title     string
	createdAt time.Time
	doneAt    null.Time
}

func NewTodoItem(title string) (TodoItem, error) {
	if err := validateTitle(title); err != nil {
		return TodoItem{}, fmt.Errorf("%w: %v", ErrInvalidTitle, err)
	}
	return TodoItem{
		id:        ulid.Make(),
		title:     title,
		createdAt: time.Now(),
	}, nil
}

func (item *TodoItem) IsDone() bool {
	return item.doneAt.Valid && !item.doneAt.Time.Before(item.createdAt)
}

func (item *TodoItem) MakeDone() error {
	if item.IsDone() {
		return ErrIsDone
	}
	item.doneAt = null.TimeFrom(time.Now())
	return nil
}

func (item *TodoItem) GetID() ulid.ULID        { return item.id }
func (item *TodoItem) GetTitle() string        { return item.title }
func (item *TodoItem) GetCreatedAt() time.Time { return item.createdAt }
func (item *TodoItem) GetDoneAt() null.Time    { return item.doneAt }

func (item *TodoItem) SetID(id ulid.ULID) {
	item.id = id
}

func (item *TodoItem) SetTitle(title string) error {
	if err := validateTitle(title); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTitle, err)
	}
	item.title = title
	return nil
}

func (item *TodoItem) SetCreatedAt(createdAt time.Time) {
	item.createdAt = createdAt
}

func (item *TodoItem) SetDoneAt(doneAt null.Time) {
	item.doneAt = doneAt
}
