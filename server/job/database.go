package job

import (
	"context"
	"database/sql"
	"server/model"
	"strings"
)

// DBRequest and DBResult are used to send SQL queries
// to the DBServer and receive results.
type DBRequest struct {
	Ctx context.Context
	Fn  func(context.Context, *sql.DB) (any, error)
	Ret chan DBResult
}

type DBResult struct {
	Value any
	Err   error
}

// Server with DB and chan
type DBServer struct {
	DB *sql.DB
	C  chan DBRequest
}

// Create a DBServer
func NewDBServer(db *sql.DB) *DBServer {
	return &DBServer{
		DB: db,
		C:  make(chan DBRequest, 64),
	}
}

// Run the DBServer
func (s *DBServer) Run(ctx context.Context) {
	for {
		select {
			// done
			case <-ctx.Done():
				return
			
			// request
			case req := <-s.C:
				value, err := req.Fn(req.Ctx, s.DB)
				select {
					// result
					case req.Ret <- DBResult{Value: value, Err: err}:
					
					// done
					case <-ctx.Done():
						return
					}
		}
	}
}

// ScanFunc is a function that scans a SQL row into a type T.
type ScanFunc[T any] func(model.Scanner) (T, error)

// WithDB sends one SQL function to the DBServer and waits for its result.
func WithDB[T any](ctx context.Context, dbChan chan<- DBRequest, fn func(context.Context, *sql.DB) (T, error)) (T, error) {
	var none T
	ret := make(chan DBResult)

	select {
	case dbChan <- DBRequest{
		Ctx: ctx,
		Fn: func(ctx context.Context, db *sql.DB) (any, error) {
			return fn(ctx, db)
		},
		Ret: ret,
	}:
	case <-ctx.Done():
		return none, ctx.Err()
	}

	select {
	case result := <-ret:
		if result.Err != nil {
			return none, result.Err
		}
		value, ok := result.Value.(T)
		if !ok {
			return none, sql.ErrNoRows
		}
		return value, nil
	case <-ctx.Done():
		return none, ctx.Err()
	}
}

// Execute a query, then scan result into a list of T.
func SqlScanList[T any](ctx context.Context, dbChan chan<- DBRequest, query string, scan ScanFunc[T], args ...any) ([]T, error) {
	return WithDB(ctx, dbChan, func(ctx context.Context, db *sql.DB) ([]T, error) {
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var results []T
		for rows.Next() {
			item, err := scan(rows)
			if err != nil {
				return nil, err
			}
			results = append(results, item)
		}
		return results, rows.Err()
	})
}

// Execute a query, then scan result into a single T.
func SqlScanOne[T any](ctx context.Context, dbChan chan<- DBRequest, query string, scan ScanFunc[T], args ...any) (T, error) {
	return WithDB(ctx, dbChan, func(ctx context.Context, db *sql.DB) (T, error) {
		row := db.QueryRowContext(ctx, query, args...)
		return scan(row)
	})
}

// Execute a query without result, but return error.
func SqlExec(ctx context.Context, dbChan chan<- DBRequest, query string, args ...any) error {
	_, err := WithDB(ctx, dbChan, func(ctx context.Context, db *sql.DB) (bool, error) {
		_, err := db.ExecContext(ctx, query, args...)
		return true, err
	})
	return err
}

// Transform a string into a SQL search pattern.
func SqlLikeSearch(search string) string {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return ""
	}
	return "%" + search + "%"
}
