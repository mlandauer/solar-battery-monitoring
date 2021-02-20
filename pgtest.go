package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Only try to read .env file if it exists
	_, err := os.Stat(".env")
	if err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Migration time
	m, err := migrate.New(
		"file://migrations",
		"postgres://postgres@localhost:5432/solar?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	err = m.Up()
	if err != nil {
		log.Fatal(err)
	}

	// Open a connection to the postgresql database "solar"
	db, err := sql.Open("pgx", "postgres://postgres@localhost:5432/solar")
	if err != nil {
		log.Fatal(err)
	}

	time := time.Now()
	battery_voltage := 2.4

	_, err = db.Exec(
		"INSERT INTO measurements (time, battery_voltage) VALUES ($1, $2)",
		time,
		battery_voltage,
	)
	if err != nil {
		log.Fatal(err)
	}
}
