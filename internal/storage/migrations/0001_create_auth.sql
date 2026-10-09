-- Baseline for the auth table. IF NOT EXISTS lets this run against databases
-- that already have the table from the old initial.sql.
CREATE TABLE IF NOT EXISTS auth (
  id serial PRIMARY KEY,
  username varchar(50) UNIQUE NOT NULL,
  password text NOT NULL
);
