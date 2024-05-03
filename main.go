package main

import (
	"context"
	"fmt"

	fooDataLayer "github.com/niko-dunixi/gorm-sample/repositories/foo"
	fooService "github.com/niko-dunixi/gorm-sample/services/foo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize root logger
	log := zerolog.New(zerolog.NewConsoleWriter())
	ctx := log.WithContext(context.Background())
	// perform dependency injection here
	db, err := initializeGorm("localhost", 5432, "user", "pass", "user")
	if err != nil {
		log.Fatal().Err(err).Msg("gorm could not be initialized")
	}
	service := fooService.NewFooService(db, fooDataLayer.NewGormFooRepo)
	// Setup any handlers here, as normal
	fooChan := make(chan fooService.Foo)
	go func() {
		defer close(fooChan)
		fooChan <- fooService.Foo{
			Name: "Foo A",
		}
		fooChan <- fooService.Foo{
			Name: "Foo B",
		}
		fooChan <- fooService.Foo{
			Name: "Foo C",
		}
	}()
	if err := service.CreateMultiple(ctx, fooChan); err != nil {
		log.Fatal().Err(err).Msg("could not create my foo items")
	}
}

func initializeGorm(host string, port int, user, pass, db string) (*gorm.DB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, db,
	)
	psql := postgres.Open(psqlInfo)
	gormDB, err := gorm.Open(psql, &gorm.Config{})
	if err != nil {
		log.Err(err).Msg("couldnt open postgres connection")
		return nil, err
	}
	return gormDB, nil
}
