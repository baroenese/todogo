package todo

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool       *pgxpool.Pool
	ErrNilPool = errors.New("cannot assign nill pool")
)

func SetPool(newPool *pgxpool.Pool) error {
	if newPool == nil {
		return ErrNilPool
	}
	pool = newPool
	return nil
}
