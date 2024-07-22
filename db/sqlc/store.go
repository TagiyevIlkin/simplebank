package db

import (
	"context"

	_ "github.com/golang/mock/mockgen/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides all functions to exxecute db queries & transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TrasferTxParams) (TransferTxResult, error)
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
}

// SQLStore provides all functions to exxecute SQL queries & transactions
type SQLStore struct {
	// Composition
	connPool *pgxpool.Pool
	*Queries
}

// NewStore creates a new Store
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}
