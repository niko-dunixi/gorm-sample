package foo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/niko-dunixi/gorm-sample/gormcommon"
	dataAccessLayer "github.com/niko-dunixi/gorm-sample/repositories/foo"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

/*
----
Here are the base types that you want to export.
Depend on these types for abstraction
----
*/

// Separates the data-access layer
type FooService interface {
	CreateMultipleAtomic(ctx context.Context, foo <-chan Foo) error
	CreateMultiple(ctx context.Context, foo <-chan Foo) error
}

// API level type meant to decouple REST from database implementation.
// Now you can refactor the database arbitrarily and more painlessly
type Foo struct {
	ID   uuid.UUID `gorm:"id"`
	Name string    `gorm:"name"`
}

/*
----
Here are the implementations. You'll need these to actually construct
higher-up before injecting them; but business logic should not know about
or depend on them.
----
*/

func NewFooService(db *gorm.DB, fooRepoProvider dataAccessLayer.FooProvider) FooService {
	return &gormFooService{
		db: db,
	}
}

type gormFooService struct {
	db           *gorm.DB
	repoProvider dataAccessLayer.FooProvider
}

func (service *gormFooService) CreateMultipleAtomic(ctx context.Context, fooChan <-chan Foo) error {
	// Start a transaction
	return gormcommon.InTx(ctx, service.db, func(tx *gorm.DB) error {
		log := zerolog.Ctx(ctx)
		ctx, cancelCtx := context.WithCancel(ctx)
		defer cancelCtx()

		repo := service.repoProvider(tx)

		for {
			select {
			case <-ctx.Done():
				err := ctx.Err()
				log.Error().Err(err).Msg("context was killed while creating many")
				return fmt.Errorf("could not complete creation of many foo: %w", err)
			case apiFoo := <-fooChan:
				dataLayerFoo := MapFooApiToDataLayer(apiFoo)
				if err := repo.Create(ctx, &dataLayerFoo); err != nil {
					return fmt.Errorf("could not create a foo: %w", err)
				}
			}
		}
	})
}

func (service *gormFooService) CreateMultiple(ctx context.Context, fooChan <-chan Foo) error {
	// Start a transaction
	log := zerolog.Ctx(ctx)
	repo := service.repoProvider(service.db)
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			log.Error().Err(err).Msg("context was killed while creating many")
			return fmt.Errorf("could not complete creation of many foo: %w", err)
		case apiFoo := <-fooChan:
			dataLayerFoo := MapFooApiToDataLayer(apiFoo)
			if err := repo.Create(ctx, &dataLayerFoo); err != nil {
				return fmt.Errorf("could not create a foo: %w", err)
			}
		}
	}
}

// The functions that convert between types are placed at the layer at which
// promotes uni-directional knowledge (EG: these are at the service layer and
// not the data access layer.)

func MapFooApiToDataLayer(foo Foo) dataAccessLayer.Foo {
	return dataAccessLayer.Foo{
		ID:   foo.ID,
		Name: foo.Name,
	}
}

func MapDataLayerToApi(foo dataAccessLayer.Foo) Foo {
	return Foo{
		ID:   foo.ID,
		Name: foo.Name,
	}
}
