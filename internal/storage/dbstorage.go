package storage

import (
	"context"
	"database/sql"
	"errors"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func (d *DBStorage) Add(ctx context.Context, userID, shortURL, originURL string) error {
	res, err := d.db.ExecContext(ctx, "INSERT INTO urls(original_url, short_url) VALUES ($1, $2) ON CONFLICT (original_url) DO NOTHING", originURL, shortURL)
	if err != nil {
		return err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return err
	} else if row < 1 {
		return errors.New("not unique original_url")
	}
	_, err = d.db.ExecContext(ctx, "INSERT INTO users(user_id, short_url) VALUES ($1, $2)", userID, shortURL)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBStorage) Get(ctx context.Context, shortURL string) (string, error) {
	var originalURL string
	row := d.db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url = $1", shortURL)
	err := row.Scan(&originalURL)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func (d *DBStorage) GetByUser(ctx context.Context, userID string) (map[string]string, error) {
	rst := make(map[string]string)
	rows, err := d.db.QueryContext(ctx, "SELECT t2.short_url, t2.original_url FROM users AS t1 LEFT JOIN urls AS t2 ON t1.short_url = t2.short_url WHERE t1.user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL string
		var origURL string
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

func (d *DBStorage) CheckStorage(ctx context.Context) error {
	err := d.db.PingContext(ctx)
	if err != nil {
		return err
	}
	var count int
	row := d.db.QueryRowContext(ctx, "SELECT COUNT(*) AS count FROM users")
	if err = row.Scan(&count); err != nil {
		_, err = d.db.ExecContext(ctx, "CREATE TABLE users (user_id VARCHAR(255) PRIMARY KEY, short_url VARCHAR(255))")
		if err != nil {
			return err
		}
	}
	row = d.db.QueryRowContext(ctx, "SELECT COUNT(*) AS count FROM urls")
	if err = row.Scan(&count); err != nil {
		_, err = d.db.ExecContext(ctx, "CREATE TABLE urls (original_url TEXT PRIMARY KEY, short_url VARCHAR(255))")
		if err != nil {
			return err
		}
		_, err = d.db.ExecContext(ctx, "CREATE UNIQUE INDEX original_url_idx ON urls (original_url)")
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DBStorage) Close() error {
	return d.db.Close()
}
