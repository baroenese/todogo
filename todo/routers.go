package todo

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	sq "github.com/Masterminds/squirrel"
)

func Router() *chi.Mux {
	app := chi.NewMux()
	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var j struct {
			Msg string `json:"message"`
		}
		users := sq.Select("*").From("users").Join("emails USING (email_id)")
		active := users.Where(sq.Eq{"deleted_at": nil})
		sql, _, _ := active.ToSql()
		j.Msg = sql
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(j)
	})
	return app
}
