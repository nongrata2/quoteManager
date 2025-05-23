package repositories_test

import (
	"context"
	stdErrors "errors"
	"io"
	"log/slog"
	"regexp"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"quotemanager/internal/models"
	"quotemanager/internal/repositories"
	"quotemanager/pkg/errors"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestDB_AddQuote(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := newTestLogger()
	r := &repositories.DB{
		Log:  logger,
		Conn: mock,
	}

	type args struct {
		ctx   context.Context
		quote models.Quote
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				ctx: context.Background(),
				quote: models.Quote{
					Author: "Test Author",
					Quote:  "Test Quote",
				},
			},
			mockBehavior: func(args args) {
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO quotes (author, quote) VALUES ($1, $2)`)).
					WithArgs(args.quote.Author, args.quote.Quote).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "Error adding",
			args: args{
				ctx: context.Background(),
				quote: models.Quote{
					Author: "Test Author",
					Quote:  "Test Quote",
				},
			},
			mockBehavior: func(args args) {
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO quotes (author, quote) VALUES ($1, $2)`)).
					WithArgs(args.quote.Author, args.quote.Quote).
					WillReturnError(stdErrors.New("db insert error"))
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			err := r.AddQuote(testCase.args.ctx, testCase.args.quote)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet(), "mock expectations not met")
		})
	}
}

func TestDB_GetQuotes(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := newTestLogger()
	r := &repositories.DB{
		Log:  logger,
		Conn: mock,
	}

	type args struct {
		ctx     context.Context
		filters models.QuoteFilter
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		expected     []models.Quote
		wantErr      bool
	}{
		{
			name: "OK - No filters",
			args: args{
				ctx:     context.Background(),
				filters: models.QuoteFilter{},
			},
			mockBehavior: func(args args) {
				rows := pgxmock.NewRows([]string{"id", "author", "quote"}).
					AddRow(1, "Author1", "Quote1").
					AddRow(2, "Author2", "Quote2")
				mock.ExpectQuery(`SELECT id, author, quote FROM quotes`).WillReturnRows(rows)
			},
			expected: []models.Quote{
				{ID: 1, Author: "Author1", Quote: "Quote1"},
				{ID: 2, Author: "Author2", Quote: "Quote2"},
			},
			wantErr: false,
		},
		{
			name: "OK - With author filter",
			args: args{
				ctx:     context.Background(),
				filters: models.QuoteFilter{Author: "Author1"},
			},
			mockBehavior: func(args args) {
				rows := pgxmock.NewRows([]string{"id", "author", "quote"}).
					AddRow(1, "Author1", "Quote1")
				mock.ExpectQuery(`SELECT id, author, quote FROM quotes WHERE author = \$1`).
					WithArgs("Author1").
					WillReturnRows(rows)
			},
			expected: []models.Quote{
				{ID: 1, Author: "Author1", Quote: "Quote1"},
			},
			wantErr: false,
		},
		{
			name: "Query Error",
			args: args{
				ctx:     context.Background(),
				filters: models.QuoteFilter{},
			},
			mockBehavior: func(args args) {
				mock.ExpectQuery(`SELECT id, author, quote FROM quotes`).
					WillReturnError(stdErrors.New("db query error"))
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Scan Error",
			args: args{
				ctx:     context.Background(),
				filters: models.QuoteFilter{},
			},
			mockBehavior: func(args args) {
				rows := pgxmock.NewRows([]string{"id", "author", "quote"}).
					AddRow("1", "Author1", "Quote1").
					RowError(0, stdErrors.New("scan error for row 0"))
				mock.ExpectQuery(`SELECT id, author, quote FROM quotes`).WillReturnRows(rows)
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {

			testCase.mockBehavior(testCase.args)

			quotes, err := r.GetQuotes(testCase.args.ctx, testCase.args.filters)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, quotes)
			}
			assert.NoError(t, mock.ExpectationsWereMet(), "mock expectations not met")
		})
	}
}
func TestDB_GetRandomQuote(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer mock.Close()

	logger := newTestLogger()
	r := &repositories.DB{
		Log:  logger,
		Conn: mock,
	}

	type args struct {
		ctx context.Context
	}

	testTable := []struct {
		name         string
		mockBehavior func()
		args         args
		expected     models.Quote
		wantErr      bool
		expectedErr  error
	}{
		{
			name: "OK",
			args: args{ctx: context.Background()},
			mockBehavior: func() {
				rows := pgxmock.NewRows([]string{"id", "quote", "author"}).
					AddRow(1, "Random Quote", "Random Author")
				mock.ExpectQuery(`SELECT id, quote, author FROM quotes ORDER BY RANDOM\(\) LIMIT 1`).
					WillReturnRows(rows)
			},
			expected:    models.Quote{ID: 1, Quote: "Random Quote", Author: "Random Author"},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "No Rows - ErrQuoteNotFound",
			args: args{ctx: context.Background()},
			mockBehavior: func() {
				mock.ExpectQuery(`SELECT id, quote, author FROM quotes ORDER BY RANDOM\(\) LIMIT 1`).
					WillReturnError(pgx.ErrNoRows)
			},
			expected:    models.Quote{},
			wantErr:     true,
			expectedErr: errors.ErrQuoteNotFound,
		},
		{
			name: "DB Error",
			args: args{ctx: context.Background()},
			mockBehavior: func() {
				mock.ExpectQuery(`SELECT id, quote, author FROM quotes ORDER BY RANDOM\(\) LIMIT 1`).
					WillReturnError(errors.ErrQuery)
			},
			expected:    models.Quote{},
			wantErr:     true,
			expectedErr: errors.ErrQuery,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior()

			actualQuote, actualErr := r.GetRandomQuote(testCase.args.ctx)

			if testCase.wantErr {
				assert.Error(t, actualErr, "Expected an error")
				if testCase.expectedErr != nil {
					if !assert.ErrorIs(t, actualErr, testCase.expectedErr) {
						assert.EqualError(t, actualErr, testCase.expectedErr.Error(), "Error message mismatch")
					}
				}
				assert.Equal(t, testCase.expected, actualQuote, "Quote data on error mismatch")
			} else {
				assert.NoError(t, actualErr, "Did not expect an error")
				assert.Equal(t, testCase.expected, actualQuote, "Quote data on success mismatch")
			}
			assert.NoError(t, mock.ExpectationsWereMet(), "mock expectations not met")
		})
	}
}

func TestDB_DeleteQuote(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer mock.Close()

	logger := newTestLogger()
	r := &repositories.DB{
		Log:  logger,
		Conn: mock,
	}

	type args struct {
		ctx     context.Context
		quoteID string
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
		expectedErr  error
	}{
		{
			name: "OK - Quote deleted",
			args: args{
				ctx:     context.Background(),
				quoteID: "id123",
			},
			mockBehavior: func(args args) {
				mock.ExpectExec(`DELETE FROM quotes WHERE id = \$1`).
					WithArgs(args.quoteID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: false,
		},
		{
			name: "Quote Not Found - ErrQuoteNotFound",
			args: args{
				ctx:     context.Background(),
				quoteID: "idNonExistent",
			},
			mockBehavior: func(args args) {
				mock.ExpectExec(`DELETE FROM quotes WHERE id = \$1`).
					WithArgs(args.quoteID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantErr:     true,
			expectedErr: errors.ErrQuoteNotFound,
		},
		{
			name: "DB Error on exec",
			args: args{
				ctx:     context.Background(),
				quoteID: "1",
			},
			mockBehavior: func(args args) {
				mock.ExpectExec(`DELETE FROM quotes WHERE id = \$1`).
					WithArgs(args.quoteID).
					WillReturnError(errors.ErrExecDB)
			},
			wantErr:     true,
			expectedErr: errors.ErrExecDB,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			actualErr := r.DeleteQuote(testCase.args.ctx, testCase.args.quoteID)

			if testCase.wantErr {
				assert.Error(t, actualErr, "Expected an error, but got nil")
				if testCase.expectedErr != nil {
					if !assert.ErrorIs(t, actualErr, testCase.expectedErr) {
						assert.EqualError(t, actualErr, testCase.expectedErr.Error(),
							"Error message mismatch (after ErrorIs failed). Actual: '%v', Expected text: '%s'",
							actualErr, testCase.expectedErr.Error())
					}
				}
			} else {
				assert.NoError(t, actualErr, "Expected no error, but got one")
			}
			assert.NoError(t, mock.ExpectationsWereMet(), "mock expectations not met")
		})
	}
}
