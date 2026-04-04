package storage

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/lib/pq"
)

func OpenPostgres() (*sql.DB, error) {
    host := getenv("DB_HOST", "postgres")
    port := getenv("DB_PORT", "5432")
    user := getenv("DB_USER", "postgres")
    pass := getenv("DB_PASSWORD", "postgres")
    name := getenv("DB_NAME", "diag")
    ssl := getenv("DB_SSLMODE", "disable")

    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, name, ssl)
    return sql.Open("postgres", dsn)
}

func getenv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}