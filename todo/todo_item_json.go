package todo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"gopkg.in/guregu/null.v4"
)

func (item TodoItem) MarshalJSON() ([]byte, error) {
	id := item.GetID()
	title := item.GetTitle()
	createdAt := item.GetCreatedAt()
	doneAt := item.GetDoneAt().Ptr()
	isDone := item.IsDone()
	j := struct {
		Id        ulid.ULID  `json:"id"`
		Title     string     `json:"title"`
		CreatedAt time.Time  `json:"created_at"`
		DoneAt    *time.Time `json:"done_at,omitempty"`
		IsDone    bool       `json:"is_done"`
	}{
		Id:        id,
		Title:     title,
		CreatedAt: createdAt,
		DoneAt:    doneAt,
		IsDone:    isDone,
	}
	return json.Marshal(j)
}

func (item *TodoItem) UnmarshalJSON(data []byte) error {
	var j struct {
		Id        ulid.ULID   `json:"id"`
		Title     string      `json:"title"`
		CreatedAt string      `json:"created_at"`
		DoneAt    null.String `json:"done_at"`
	}
	if err := json.Unmarshal(data, &j); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339, j.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to parse 'created_at': %w", err)
	}
	item.SetID(j.Id)
	if err := item.SetTitle(j.Title); err != nil {
		return fmt.Errorf("failed to set title: %w", err)
	}
	item.SetCreatedAt(createdAt)
	item.SetDoneAt(parseNullStringToNullTime(j.DoneAt))
	return nil
}

func parseNullStringToNullTime(s null.String) null.Time {
	if !s.Valid {
		return null.Time{}
	}
	ts, err := time.Parse(time.RFC3339, s.String)
	if err != nil {
		return null.Time{}
	}
	return null.TimeFrom(ts)
}
