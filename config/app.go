package config

import (
	"database/sql"
	"os"

	"github.com/safrizal-hk/uas-gofiber/database"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

type Database struct {
	PgDB    *sql.DB       
	MongoDB *mongo.Database 
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		os.Exit(1)
	}
}

func NewDB() *Database {
	return &Database{
		PgDB:    database.ConnectPostgreSQL(),
		MongoDB: database.ConnectMongoDB(),
	}
}