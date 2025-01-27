package todo

import (
	"errors"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbPool *pgxpool.Pool
	poolMu sync.RWMutex
)

var (
	ErrNilPool            = errors.New("cannot assign nill pool")
	ErrPoolNotInitialized = errors.New("database pool is not initialized")
)

func SetPool(newPool *pgxpool.Pool) error {
	if newPool == nil {
		return ErrNilPool
	}
	poolMu.Lock()
	defer poolMu.Unlock()
	dbPool = newPool
	return nil
}

func GetPool() (*pgxpool.Pool, error) {
	poolMu.RLock()
	defer poolMu.RUnlock()
	if dbPool == nil {
		return nil, ErrPoolNotInitialized
	}
	return dbPool, nil
}

func IsPoolInitialized() bool {
	poolMu.RLock()
	defer poolMu.RUnlock()
	return dbPool != nil
}
