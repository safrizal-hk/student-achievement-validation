package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB() *mongo.Database {
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")
	if mongoURI == "" || dbName == "" {
		log.Fatal("MONGO_URI atau MONGO_DB_NAME environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Gagal membuat client MongoDB: ", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("Gagal konek ke MongoDB: ", err)
	}

	log.Println("Koneksi MongoDB berhasil.")
	return client.Database(dbName)
}