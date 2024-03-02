CREATE DATABASE auth;

CREATE TABLE "auth" (
  "id" serial NOT NULL,
  PRIMARY KEY ("id"),
  "username" VARCHAR(50) UNIQUE NOT NULL,
  "password" TEXT NOT NULL
);