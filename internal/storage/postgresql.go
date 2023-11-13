package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Postgresql contains a connection to the database and the necessary methods for working with data.
type Postgresql struct {
	db *sql.DB
}

// NewPostgresql is the constructor for the Postgresql structure.
func NewPostgresql(ctx context.Context, addrConnDB string) (*Postgresql, error) {
	db, err := sql.Open("postgres", addrConnDB)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	if err := createTables(ctx, db); err != nil {
		return nil, err
	}

	return &Postgresql{db: db}, nil
}

// Add records the user's id, the original URL, and its shortened URL.
func (d *Postgresql) Add(ctx context.Context, userID, shortURL, originURL string) error {
	const op = "internal.storage.postgresql.Add"

	query := `INSERT INTO 
    			urls(original_url, short_url) 
			VALUES ($1, $2) 
			ON CONFLICT (original_url) DO NOTHING`
	res, err := d.db.ExecContext(ctx, query, originURL, shortURL)
	if err != nil {
		return fmt.Errorf("%s.InsertIntoURLs: %w", op, err)
	}

	row, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s.RowsAffected: %w", op, err)
	}

	if row < 1 {
		return ErrUniqueValue
	}

	query = `INSERT INTO 
    			users(user_id, short_url) 
			VALUES ($1, $2)`
	_, err = d.db.ExecContext(ctx, query, userID, shortURL)
	if err != nil {
		return fmt.Errorf("%s.InsertIntoUsers: %w", op, err)
	}

	return nil
}

// Get retrieves the original URL from the database by its shortened value.
func (d *Postgresql) Get(ctx context.Context, shortURL string) (string, error) {
	var (
		originalURL string
		markDelete  bool
	)

	query := `SELECT 
    		original_url, 
    		COALESCE(mark_del, FALSE) AS mark_del 
		FROM urls 
		WHERE short_url = $1`
	row := d.db.QueryRowContext(ctx, query, shortURL)

	err := row.Scan(&originalURL, &markDelete)
	if err != nil {
		return "", err
	} else if markDelete {
		return "", ErrDeletedURL
	} else {
		return originalURL, nil
	}
}

// GetByUser receives a map of shortened and original URLs by user from the database.
func (d *Postgresql) GetByUser(ctx context.Context, userID string) (map[string]string, error) {
	rst := make(map[string]string)

	query := `SELECT 
    		t2.short_url, 
    		t2.original_url 
		FROM 
		    users AS t1 
		    	LEFT JOIN urls AS t2 
		    	ON t1.short_url = t2.short_url 
		WHERE 
		    t1.user_id = $1`

	rows, err := d.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			shortURL string
			origURL  string
		)

		err = rows.Scan(&shortURL, &origURL)
		if err != nil {
			return nil, err
		}

		rst[shortURL] = origURL
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return rst, nil
}

// Delete marks the shortened URL in the database as deleted.
func (d *Postgresql) Delete(ctx context.Context, shortURL string) error {
	query := `UPDATE urls 
		SET mark_del = TRUE 
		WHERE short_url = $1`

	_, err := d.db.ExecContext(ctx, query, shortURL)
	if err != nil {
		return err
	}

	return nil
}

// CheckStorage checks the connection to the database.
func (d *Postgresql) CheckStorage(ctx context.Context) error {
	err := d.db.PingContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetStats selects data for statistics from the database.
func (d *Postgresql) GetStats(ctx context.Context) (int, int) {
	var (
		countURLs  int
		countUsers int
	)
	query :=
		`SELECT
			COUNT(user_id) AS countUsers,
			COUNT(short_url) AS countURLs
		FROM users`

	row := d.db.QueryRowContext(ctx, query)
	err := row.Scan(&countUsers, &countURLs)
	if err != nil {
		return 0, 0
	}

	return countURLs, countUsers
}

// Close closes the connection to the database.
func (d *Postgresql) Close() error {
	return d.db.Close()
}

func createTables(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
    		user_id VARCHAR(255), 
    		short_url VARCHAR(255) PRIMARY KEY);
		CREATE INDEX IF NOT EXISTS idx_user ON users(user_id);
		CREATE INDEX IF NOT EXISTS idx_url ON users(short_url);`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = `
		CREATE TABLE IF NOT EXISTS urls (
    		original_url TEXT PRIMARY KEY, 
    		short_url VARCHAR(255), 
    		mark_del BOOLEAN);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls(original_url)`

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
