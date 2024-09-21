package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/hasanhakkaev/yqapp-demo/assets"
	"time"
)

type Postgres struct {
	DB *sql.DB
}

func NewPostgres(dsn string, autoMigrate bool) (*Postgres, error) {

	db, err := sql.Open("postgres", "postgres://"+dsn+"?sslmode=disable")
	if err != nil {
		return nil, err
	}

	// Try connecting to the database
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	timeoutExceeded := time.After(60 * time.Second)
	for {
		select {
		case <-timeoutExceeded:
			return nil, fmt.Errorf("connection timeout")

		case <-ticker.C:
			if err := db.Ping(); err == nil {

				db.SetMaxOpenConns(25)
				db.SetMaxIdleConns(25)
				db.SetConnMaxIdleTime(5 * time.Minute)
				db.SetConnMaxLifetime(2 * time.Hour)

				if autoMigrate {
					fsDriver, err := iofs.New(assets.EmbeddedFiles, "migrations")
					if err != nil {
						return nil, err
					}

					migrator, err := migrate.NewWithSourceInstance("iofs", fsDriver, "postgres://"+dsn+"?sslmode=disable")
					if err != nil {
						return nil, err
					}

					err = migrator.Up()
					switch {
					case errors.Is(err, migrate.ErrNoChange):
						break
					case err != nil:
						return nil, err
					}
				}

				return &Postgres{DB: db}, nil
			}
		}
	}

}
