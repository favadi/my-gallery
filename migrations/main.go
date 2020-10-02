package main

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose"
)

const driver = "postgres"

func main() {
	var dbString = flag.String("db-dsn", "postgres://my-gallery:my-gallery@127.0.0.1:5432/my-gallery?sslmode=disable", "PostgresQL database DSN")

	flag.Parse()

	args := flag.Args()

	db, err := sql.Open(driver, *dbString)
	if err != nil {
		log.Fatal(err)
	}

	if err = goose.SetDialect(driver); err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		log.Fatal("expected at least one arg")
	}

	command := args[0]

	if err = goose.Run(command, db, ".", args[1:]...); err != nil {
		log.Fatalf("goose run: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}
