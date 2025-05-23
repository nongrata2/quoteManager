package repositories

import (
	"errors"
	"fmt"
	"quotemanager/migrations"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func (db *DB) Migrate() error {
	db.Log.Debug("running migration")

	pool, ok := db.Conn.(*pgxpool.Pool)
	if !ok {
		err := errors.New("db.Conn is not of type *pgxpool.Pool, cannot run migrations")
		db.Log.Error("type assertion failed for db.Conn to *pgxpool.Pool", "actual_type", fmt.Sprintf("%T", db.Conn), "error", err)
		return err
	}

	files, err := iofs.New(migrations.MigrationFiles, ".")
	if err != nil {
		db.Log.Error("failed to load migration files", "error", err)
		return err
	}
	db.Log.Debug("migration files loaded successfully")

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	driver, err := pgx.WithInstance(sqlDB, &pgx.Config{})
	if err != nil {
		db.Log.Error("failed to create pgx driver for migrations", "error", err)
		return err
	}
	m, err := migrate.NewWithInstance("iofs", files, "pgx", driver)
	if err != nil {
		db.Log.Error("failed to initialize migrations", "error", err)
		return err
	}

	err = m.Up()

	if err != nil {
		if err != migrate.ErrNoChange {
			db.Log.Error("migration failed", "error", err)
			return err
		}
		db.Log.Debug("migration did not change anything")
	}

	db.Log.Debug("migration finished")
	return nil
}
