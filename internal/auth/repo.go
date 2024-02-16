package auth

import (
	"fmt"
	"html"

	"github.com/vilmis04/auth-proxy/internal/storage"
)

type Repo struct {
	storage.Storage
}

func NewRepo() *Repo {
	return &Repo{
		Storage: *storage.New("auth"),
	}
}

func (r *Repo) GetUserList() (*[]string, error) {
	db, err := r.ConnectToDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := fmt.Sprintf(`SELECT name FROM %v`, r.Table)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}

		names = append(names, name)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &names, nil
}

func (r *Repo) createUser(body signUpRequest) error {
	db, err := r.ConnectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf(`
	INSERT INTO %v
	VALUES username=$1, password=$2`, r.Table)
	_, err = db.Exec(query, html.EscapeString(body.Username), html.EscapeString(body.Password))
	if err != nil {
		return err
	}

	return nil
}
