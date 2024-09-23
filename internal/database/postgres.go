package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/hasanhakkaev/yqapp-demo/assets"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"os"
)

type Postgres struct {
	DB *pgxpool.Pool
}

var ctx = context.Background()

func NewPostgres(dsn string) (*Postgres, error) {

	db, err := pgxpool.New(ctx, "postgres://"+dsn+"?sslmode=disable")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Set the search path for the current session
	//_, err = db.Exec(ctx, "SET search_path TO yqapp_demo_schema")
	//if err != nil {
	//	log.Fatal(err)
	//}

	return &Postgres{
		DB: db,
	}, nil
}

// MigrateModels migrates the domain models using the given DB connection.
func MigrateModels(dsn string) error {

	fsDriver, err := iofs.New(assets.EmbeddedFiles, "migrations")
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", fsDriver, "postgres://"+dsn+"?sslmode=disable")
	if err != nil {
		return err
	}

	err = migrator.Up()
	switch {
	case errors.Is(err, migrate.ErrNoChange):
		break
	case err != nil:
		return err
	}
	return nil
}
