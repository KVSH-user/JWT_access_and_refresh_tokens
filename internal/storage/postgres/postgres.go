package postgres

import (
	"auth/internal/entity"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Storage struct {
	db *sql.DB
}

func New(host, port, user, password, dbname string) (*Storage, error) {
	const op = "storage.postgres.New"

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	storage := &Storage{db: db}

	err = goose.Up(storage.db, "internal/storage/migrations")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return storage, nil
}

func (s *Storage) SaveToken(user *entity.User) error {
	const op = "storage.postgres.SaveToken"

	_, err := s.db.Exec("UPDATE users SET is_valid = FALSE WHERE guid = $1", user.Guid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := s.db.Prepare("INSERT INTO users (refresh_token, guid) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.RefreshToken, user.Guid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) ValidateToken(refreshToken string) (string, error) {
	const op = "storage.postgres.ValidateToken"

	guid := ""

	stmt, err := s.db.Prepare("SELECT guid FROM users WHERE refresh_token = $1 AND is_valid = TRUE")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(refreshToken).Scan(&guid)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return guid, nil
}
