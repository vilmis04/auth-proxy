package storage

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestListMigrationsIsSortedAndFiltered(t *testing.T) {
	fsys := fstest.MapFS{
		"m/0002_b.sql": {},
		"m/0001_a.sql": {},
		"m/notes.md":   {},
	}
	got, err := listMigrations(fsys, "m")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"0001_a.sql", "0002_b.sql"}; !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestEmbeddedMigrationsExist(t *testing.T) {
	names, err := listMigrations(migrationFiles, "migrations")
	if err != nil || len(names) == 0 {
		t.Fatalf("expected embedded migrations, got %v (%v)", names, err)
	}
}

// Runs against a real database when TEST_DATABASE_URL is set.
func TestMigrateIsIdempotentAndBaselinesExistingTable(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	reset := func() {
		db.ExecContext(ctx, `DROP TABLE IF EXISTS auth, schema_migrations`)
	}
	reset()
	defer reset()

	// Existing production database: the table is already there.
	if _, err := db.ExecContext(ctx, `CREATE TABLE auth (id serial PRIMARY KEY, username varchar(50) UNIQUE NOT NULL, password text NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO auth (username, password) VALUES ('a&b', 'x')`); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := Migrate(ctx, db); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}
	}

	var rows, applied int
	db.QueryRowContext(ctx, `SELECT count(*) FROM auth`).Scan(&rows)
	db.QueryRowContext(ctx, `SELECT count(*) FROM schema_migrations`).Scan(&applied)
	if rows != 1 || applied != 1 {
		t.Errorf("expected 1 row preserved and 1 migration recorded, got %d and %d", rows, applied)
	}
}
