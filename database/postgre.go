package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func ConnectPostgreSQL() *sql.DB {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		log.Fatal("PG_DSN environment variable not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Gagal membuka koneksi PostgreSQL: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Gagal konek ke PostgreSQL: ", err)
	}

	log.Println("Koneksi PostgreSQL berhasil.")
	return db
}