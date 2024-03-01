-- Create auth table
CREATE DATABASE IF NOT EXISTS "auth_proxy";

CREATE TABLE IF NOT EXISTS "auth" (
  "id" serial NOT NULL,
  PRIMARY KEY ("id"),
  "username" character varying NOT NULL,
  "password" bit varying NOT NULL
);