package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound
	ErrDuplicatedKey  = gorm.ErrDuplicatedKey
)

// BaseRepo[T] is a generic repository that provides standard CRUD operations
// and transaction helpers for any GORM model T.
// Every method accepts a context.Context so the caller controls timeouts
// and cancellation.
type BaseRepo[T any] struct {
	db *gorm.DB
}

// NewBaseRepo creates a BaseRepo for type T backed by the given *gorm.DB.
func NewBaseRepo[T any](db *gorm.DB) BaseRepo[T] {
	return BaseRepo[T]{db: db}
}

// DB returns the underlying *gorm.DB so that concrete repositories can build
// custom scoped queries (Preload, Where, Order, WithContext, …).
func (r *BaseRepo[T]) DB() *gorm.DB {
	return r.db
}

type DB = gorm.DB

// TxFunc is the callback passed to Transaction/TransactionCtx.
// The provided tx already carries the context supplied to Transaction.
type TxFunc = func(tx *gorm.DB) error

// -----------------------------------------------------------------
// Transaction
// -----------------------------------------------------------------

// Transaction wraps fn in a database transaction scoped to ctx.
// If fn returns an error the transaction is rolled back automatically.
func (r *BaseRepo[T]) Transaction(ctx context.Context, fn TxFunc) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// -----------------------------------------------------------------
// Standard CRUD (all context-aware)
// -----------------------------------------------------------------

func (r *BaseRepo[T]) FindAll(ctx context.Context) ([]*T, error) {
	var records []*T
	err := r.db.WithContext(ctx).Find(&records).Error
	return records, err
}

func (r *BaseRepo[T]) FindByID(ctx context.Context, id int) (*T, error) {
	var record T
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *BaseRepo[T]) Create(ctx context.Context, record *T) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *BaseRepo[T]) Save(ctx context.Context, record *T) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *BaseRepo[T]) Update(ctx context.Context, column string, record *T) error {
	return r.db.WithContext(ctx).Update(column, record).Error
}

func (r *BaseRepo[T]) Delete(ctx context.Context, record *T) error {
	return r.db.WithContext(ctx).Delete(record).Error
}

// -----------------------------------------------------------------
// Scoped queries (caller provides a pre-configured *gorm.DB)
// Use r.DB().WithContext(ctx).Preload(…).Where(…) to build the scope.
// -----------------------------------------------------------------

func (r *BaseRepo[T]) FindAllScoped(scope *gorm.DB) ([]*T, error) {
	var records []*T
	if err := scope.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *BaseRepo[T]) FirstScoped(scope *gorm.DB) (*T, error) {
	var record T
	if err := scope.First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// -----------------------------------------------------------------
// Transaction-aware CRUD helpers
// (tx already carries the context from Transaction)
// -----------------------------------------------------------------

func (r *BaseRepo[T]) CreateTx(record *T, tx *gorm.DB) error {
	return tx.Create(record).Error
}

func (r *BaseRepo[T]) SaveTx(record *T, tx *gorm.DB) error {
	return tx.Save(record).Error
}

func (r *BaseRepo[T]) UpdateTx(column string, record *T, tx *gorm.DB) error {
	return tx.Update(column, record).Error
}

func (r *BaseRepo[T]) DeleteTx(record *T, tx *gorm.DB) error {
	return tx.Delete(record).Error
}

// UpsertTx performs INSERT … ON CONFLICT (conflictCols) DO UPDATE (updateCols)
// within an existing transaction.
func (r *BaseRepo[T]) UpsertTx(record *T, conflictCols []string, updateCols []string, tx *gorm.DB) error {
	cols := make([]clause.Column, len(conflictCols))
	for i, c := range conflictCols {
		cols[i] = clause.Column{Name: c}
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   cols,
		DoUpdates: clause.AssignmentColumns(updateCols),
	}).Create(record).Error
}
