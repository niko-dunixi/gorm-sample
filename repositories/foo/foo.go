package foo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

/*
----
Here are the base types that you want to export.
Depend on these types for abstraction
----
*/

// Used for dependency injection
type FooProvider func(*gorm.DB) FooRepository

// Separates the data-access layer
type FooRepository interface {
	Create(ctx context.Context, foo *Foo) error
	GetByID(ctx context.Context, id uuid.UUID) (*Foo, error)
}

// Database level type
type Foo struct {
	ID   uuid.UUID `gorm:"id"`
	Name string    `gorm:"name"`
}

// A custom sentinal error that gives business logic more flexibilty
type ErrNoFooByID struct {
	ID uuid.UUID
}

func (err ErrNoFooByID) Error() string {
	return fmt.Sprintf("there was no foo with id: %s", err.ID)
}

/*
----
Here are the implementations. You'll need these to actually construct
higher-up before injecting them; but business logic should not know about
or depend on them.
----
*/

func NewGormFooRepo(db *gorm.DB) FooRepository {
	if db == nil {
		// This is called by business logic, it doesn't make sense
		// for business logic to attempt to recover from an improper
		// setup. Guard for this properly where you do dependency
		// injection
		panic("")
	}
	return &gormFooRepo{
		db: db,
	}
}

// Since we are masking the repository with an interface, we name
// the implementation in a way to delimit it's specificity. This
// allows us to potentially migrate or support multiple implementations
// later and know which one does which.
type gormFooRepo struct {
	db *gorm.DB
}

func (repo *gormFooRepo) Create(ctx context.Context, foo *Foo) error {
	if id := foo.ID; id != uuid.Nil {
		// Since we're doing a creation, we must guard against
		// receiving an entity that has already been persisted
		return fmt.Errorf(
			"create was called with a foo that already had an id: %s",
			id,
		)
	}
	foo.ID = uuid.New()
	// Return the error straight back from the attempt to create
	// there aren't any sensible cases we want to operate upon
	// conditionally in business logic (all errors are unexpected errors)
	return repo.db.WithContext(ctx).Create(foo).Error
}

func (repo *gormFooRepo) GetByID(ctx context.Context, id uuid.UUID) (*Foo, error) {
	result := Foo{}
	err := repo.db.WithContext(ctx).
		Take(&result, "id = ?", id).
		Error
	// Check for errors, if we know this is a scenario that we want to manage
	// in business logic, return a custom error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoFooByID{
			ID: id,
		}
	} else if err != nil {
		// We still want to manage unexpected errors that we cannot recover from
		return nil, fmt.Errorf("could not locate store type by id (%s): %w", id, err)
	}
	// Happy path
	return &result, nil
}
