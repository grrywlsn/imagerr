package db

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    _ "github.com/lib/pq"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

var DB *sql.DB

func InitDB() {
    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
    )

    var err error
    DB, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Error connecting to database:", err)
    }

    err = DB.Ping()
    if err != nil {
        log.Fatal("Error pinging database:", err)
    }

    // Run migrations
    if err := runMigrations(DB); err != nil {
        log.Fatal("Error running migrations:", err)
    }

    log.Println("Successfully connected to database and ran migrations")
}

func runMigrations(db *sql.DB) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return fmt.Errorf("could not create postgres driver: %v", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file:///Users/garrywilson/Developer/imagerr/src/db/migrations",
        "postgres", driver)
    if err != nil {
        return fmt.Errorf("could not create migrate instance: %v", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("could not run migrations: %v", err)
    }

    return nil
}

func CloseDB() {
    if DB != nil {
        DB.Close()
    }
}