package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// pgUniqueViolation is the Postgres error code for unique_violation.
const pgUniqueViolation = "23505"

const usersTable = "auth"

type Repo struct {
	db *sql.DB
}

// NewRepo uses the shared pool; it never opens or closes connections itself.
func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// CreateUser stores the username exactly as given. The UNIQUE constraint on
// the column is the source of truth for duplicate detection.
func (r *Repo) CreateUser(username string, passwordHash []byte) error {
	query := fmt.Sprintf(`
	INSERT INTO %v (username, password)
	VALUES ($1, $2)`, usersTable)
	_, err := r.db.Exec(query, username, passwordHash)
	var pgErr *pq.Error
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return ErrUsernameTaken
	}

	return err
}

func (r *Repo) GetUser(username string) (*User, error) {
	query := fmt.Sprintf(`
	SELECT username, password FROM %v
	WHERE username=$1`, usersTable)

	user := User{}
	err := r.db.QueryRow(query, username).Scan(&user.Username, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
