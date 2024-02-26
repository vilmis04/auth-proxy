package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Storage struct {
	Table      string
	connString string
}

func New(table string) *Storage {
	return &Storage{
		Table: table,
		connString: fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT"),
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB")),
	}
}

func (s *Storage) ConnectToDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", s.connString)
	if err != nil {
		return nil, err
	}

	return db, nil
}
