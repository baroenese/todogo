package todo

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"
)

var (
	ErrTodoNotFound = errors.New("todo: not found")
)

func Router() *chi.Mux {
	app := chi.NewMux()
	app.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			h.ServeHTTP(w, r)
		})
	})
	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		resp, err := listItems(ctx)
		if err != nil {
			writeMessageError(w, http.StatusInternalServerError, err.Error())
			return
		}
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
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})
	return app
}

func writeMessageError(w http.ResponseWriter, status int, msg string) {
	var j struct {
		Msg string `json:"message"`
	}
	j.Msg = msg
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(j)
}
