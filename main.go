package main

import (
	"log"
	"os"
	
	"github.com/gofiber/fiber/v2"
	
	"github.com/safrizal-hk/uas-gofiber/config"
	"github.com/safrizal-hk/uas-gofiber/route" 
)

func main() {
	config.LoadEnv()

	dbConn := config.NewDB()
	
	defer dbConn.PgDB.Close()
	
	app := fiber.New() 

	route.RegisterAllRoutes(app, dbConn)
	
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}