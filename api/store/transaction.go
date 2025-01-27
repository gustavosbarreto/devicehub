package store

import (
	"context"
	"errors"
)

var ErrStartTransactionFailed = errors.New("start transaction failed")

// TransactionCb defines the function signature expected for transaction operations.
// It typically encompasses a series of store method calls that must be executed within a transaction.
type TransactionCb func(ctx context.Context) error

type TransactionStore interface {
	// WithTransaction executes a callback cb within a transaction, ensuring that a series of store
	// operations are executed as a single unit, committing the changes when the callback returns nil.
	// If any operation fails, the transaction is aborted, rolling back all operations and returning the
	// error from the callback. It returns ErrTransactionFailed if the transaction cannot start.
	WithTransaction(ctx context.Context, cb TransactionCb) error
}
