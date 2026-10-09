package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/vilmis04/auth-proxy/internal/storage"
)

// pgUniqueViolation is the Postgres error code for unique_violation.
const pgUniqueViolation = "23505"

type Repo struct {
	storage.Storage
}

func NewRepo() *Repo {
	return &Repo{
		Storage: *storage.New("auth"),
	}
}

// CreateUser stores the username exactly as given. The UNIQUE constraint on
// the column is the source of truth for duplicate detection.
func (r *Repo) CreateUser(username string, passwordHash []byte) error {
	db, err := r.ConnectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf(`
	INSERT INTO %v (username, password)
	VALUES ($1, $2)`, r.Table)
	_, err = db.Exec(query, username, passwordHash)
	var pgErr *pq.Error
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return ErrUsernameTaken
	}

	return err
}

func (r *Repo) GetUser(username string) (*User, error) {
	db, err := r.ConnectToDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := fmt.Sprintf(`
	SELECT username, password FROM %v
	WHERE username=$1`, r.Table)

	user := User{}
	err = db.QueryRow(query, username).Scan(&user.Username, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
