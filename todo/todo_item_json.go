package todo

import (
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"
	"gopkg.in/guregu/null.v4"
)

func (item TodoItem) MarshalJSON() ([]byte, error) {
	j := struct {
		Id        ulid.ULID  `json:"id"`
		Title     string     `json:"title"`
		CreatedAt time.Time  `json:"created_at"`
		DoneAt    *time.Time `json:"done_at,omitempty"`
		IsDone    bool       `json:"is_done"`
	}{
		Id:        item.Id,
		Title:     item.Title,
		CreatedAt: item.CreatedAt,
		DoneAt:    item.DoneAt.Ptr(),
		IsDone:    item.IsDone(),
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
		return err
	}
	createdAt, err := time.Parse(time.RFC3339, j.CreatedAt)
	if err != nil {
		return err
	}
	item.Id = j.Id
	item.Title = j.Title
	item.CreatedAt = createdAt
	item.DoneAt = parseNullStringToNullTime(j.DoneAt)
	return nil
}

func parseNullStringToNullTime(s null.String) (t null.Time) {
	if !s.Valid {
		return
	}
	ts, err := time.Parse(time.RFC3339, s.String)
	if err != nil {
		return
	}
	return null.TimeFrom(ts)
}
