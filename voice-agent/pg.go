package main

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

func NewSQLx() *sqlx.DB {
	ds := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		os.Getenv("VOICE_AGENT_DB_USER"),
		os.Getenv("VOICE_AGENT_DB_PASSWORD"),
		os.Getenv("VOICE_AGENT_DB_HOST"),
		os.Getenv("VOICE_AGENT_DB_PORT"),
		os.Getenv("VOICE_AGENT_DB_NAME"))
	db, err := sqlx.Open("postgres", ds)
	if err != nil {
		panic(err)
	}
	return db
}
