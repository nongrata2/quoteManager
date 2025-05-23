package repositories

import (
	"context"
	stdErrors "errors"
	"log/slog"
	"quotemanager/internal/models"
	"quotemanager/pkg/errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolConnector interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Ping(ctx context.Context) error
	Close()
}

type DBInterface interface {
	AddQuote(ctx context.Context, quote models.Quote) error
	GetQuotes(ctx context.Context, filters models.QuoteFilter) ([]models.Quote, error)
	GetRandomQuote(ctx context.Context) (models.Quote, error)
	DeleteQuote(ctx context.Context, quoteID string) error
}

type DB struct {
	Log  *slog.Logger
	Conn PoolConnector
}

var _ DBInterface = (*DB)(nil)

func New(log *slog.Logger, address string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Error("failed to ping database", "error", err)
		return nil, err
	}

	log.Info("successfully connected to database", "address", address)

	return &DB{
		Log:  log,
		Conn: pool,
	}, nil
}

func (db *DB) AddQuote(ctx context.Context, quote models.Quote) error {

	db.Log.Debug("started adding quote DB")

	query := `
        INSERT INTO quotes (author, quote)
        VALUES ($1, $2)
    `
	_, err := db.Conn.Exec(ctx, query,
		quote.Author,
		quote.Quote,
	)

	if err != nil {
		db.Log.Error("Failed to add quote", "error", err)
		return err
	}
	db.Log.Debug("Finished adding quote to DB")

	return nil
}

func (db *DB) GetQuotes(ctx context.Context, filters models.QuoteFilter) ([]models.Quote, error) {
	db.Log.Debug("started getting quote list DB")
	var quotes []models.Quote

	query := `
		SELECT id, author, quote
		FROM quotes
	`
	var args []any

	if filters.Author != "" {
		query += " WHERE author = $1"
		args = append(args, filters.Author)
	}

	db.Log.Debug("executing query", "query", strings.TrimSpace(query), "args", args)

	rows, err := db.Conn.Query(ctx, query, args...)
	if err != nil {
		db.Log.Error("failed to fetch quotes", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q models.Quote
		err := rows.Scan(
			&q.ID,
			&q.Author,
			&q.Quote,
		)
		if err != nil {
			db.Log.Error("failed to scan quote row", "error", err)
			return nil, err
		}
		quotes = append(quotes, q)
	}

	if err := rows.Err(); err != nil {
		db.Log.Error("error while iterating over rows", "error", err)
		return nil, err
	}

	db.Log.Debug("ended getting quote list DB")
	return quotes, nil
}

func (db *DB) GetRandomQuote(ctx context.Context) (models.Quote, error) {
	db.Log.Debug("started getting random quote DB")
	var quote models.Quote

	query := `
		SELECT id, quote, author
		FROM quotes
		ORDER BY RANDOM()
		LIMIT 1
	`

	db.Log.Debug("executing query", "query", strings.TrimSpace(query))

	err := db.Conn.QueryRow(ctx, query).Scan(
		&quote.ID,
		&quote.Quote,
		&quote.Author,
	)

	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			db.Log.Warn("no quotes was found in DB")
			return models.Quote{}, errors.ErrQuoteNotFound
		}
		db.Log.Error("failed to fetch or scan random quote", "error", err)
		return models.Quote{}, err
	}

	db.Log.Debug("ended getting random quote DB", "quote_id", quote.ID)
	return quote, nil
}

func (db *DB) DeleteQuote(ctx context.Context, quoteID string) error {
	db.Log.Debug("started deleting quote from DB")

	query := `DELETE FROM quotes WHERE id = $1`

	result, err := db.Conn.Exec(ctx, query, quoteID)
	if err != nil {
		db.Log.Error("failed to delete quote", "error", err)
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		db.Log.Warn("no quote was found with the given id", "id", quoteID)
		return errors.ErrQuoteNotFound
	}
	db.Log.Debug("Finished deleting quote from DB")
	return nil
}
