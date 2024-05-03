package gormcommon

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type AuditFields struct {
	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
	// This is a common trip up. Gorm will only respect `deleted_at`
	// if it is the type `gorm.DeletedAt`.
	DeletedAt gorm.DeletedAt `gorm:"deleted_at"`
}

type TxOperation func(tx *gorm.DB) error

func InTx(ctx context.Context, db *gorm.DB, operation TxOperation) error {
	log := zerolog.Ctx(ctx)
	// Start the transaction
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Setup base error handling. Gorm will not control panics
		// so you must if you wish to recover from a panicked transaction
		var err error
		defer func() {
			if panicValue := recover(); panicValue != nil {
				stack := string(debug.Stack())
				if panicErr, isErr := panicValue.(error); isErr {
					log.Error().Err(panicErr).Str("stack", stack).Msg("a panic occurred in the gorm transaction")
					err = fmt.Errorf("a panic ocurred in the gorm transaction: %v", panicValue)
				} else {
					log.Error().Interface("panic", panicValue).Str("stack", stack).Msg("a panic occurred in the gorm transaction")
					err = fmt.Errorf("a panic ocurred in the gorm transaction: %v", panicValue)
				}
			}
		}()
		// Call the operation that you actually want to perform
		err = operation(tx)
		return err
	})
}
