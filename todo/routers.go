package todo

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"

	sq "github.com/Masterminds/squirrel"
)

func Router() *chi.Mux {
	app := chi.NewMux()
	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		resp, err := listItems(ctx)
		if err != nil {
			writeMessageError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	app.Get("/{itemId}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		itemId := chi.URLParam(r, "itemId")
		id, err := ulid.Parse(itemId)
		if err != nil {
			writeMessageError(w, http.StatusBadRequest, err.Error())
			return
		}
		var resp TodoItem
		item, err := findItem(ctx, id)
		if err != nil {
			errStr := err.Error()
			if err == ErrTodoNotFound {
				writeMessageError(w, http.StatusNotFound, errStr)
				return
			}
			writeMessageError(w, http.StatusInternalServerError, errStr)
			return
		}
		resp = item
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	app.Post("/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			writeMessageError(w, http.StatusBadRequest, err.Error())
			return
		}
		ctx := r.Context()
		title := r.FormValue("title")
		id, err := createItem(ctx, title)
		if err != nil {
			writeMessageError(w, http.StatusBadRequest, err.Error())
			return
		}
		var resp struct {
			Id string `json:"id"`
		}
		resp.Id = id.String()
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	app.Post("/done", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			writeMessageError(w, http.StatusBadRequest, err.Error())
			return
		}
		ctx := r.Context()
		idStr := r.FormValue("id")
		id, err := ulid.Parse(idStr)
		if err != nil {
			writeMessageError(w, http.StatusBadRequest, err.Error())
			return
		}
		err = makeItemDone(ctx, id)
		if err != nil {
			writeMessageError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var resp struct {
			Id string `json:"id"`
		}

		resp.Id = id.String()
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})
	app.Get("/test-query", func(w http.ResponseWriter, r *http.Request) {
		var j struct {
			Msg string `json:"message"`
		}
		sqlCreate := sq.Insert("todolist").
			Columns("id", "title", "created_at", "done_at").
			Values("item.Id", "item.Title", "item.CreatedAt", "item.DoneAt").
			Suffix("ON CONFLICT(id) DO UPDATE SET title=$2, done_at=$4").
			PlaceholderFormat(sq.Dollar)

		sqlStrCreate, _, _ := sqlCreate.ToSql()
		j.Msg = sqlStrCreate
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(j)
	})
	return app
}

func writeMessageError(w http.ResponseWriter, status int, msg string) {
	var j struct {
		Msg string `json:"message"`
	}
	j.Msg = msg
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(j)
}
