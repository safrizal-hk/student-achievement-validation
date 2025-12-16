package main

import (
	"log"
	"os"
	
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/config"
	"github.com/safrizal-hk/uas-gofiber/route" 

	_ "github.com/safrizal-hk/uas-gofiber/docs" 
	"github.com/swaggo/fiber-swagger"
)

// http://localhost:3000/swagger/index.html

// @title           Sistem Pelaporan Prestasi Mahasiswa API
// @version         1.0
// @description     Dokumentasi API untuk UAS Backend Lanjut (Hybrid Database: PostgreSQL & MongoDB).
// @termsOfService  http://swagger.io/terms/

// @contact.name    Safrizal Huda
// @contact.email   admin@univ.ac.id

// @host            localhost:3000
// @BasePath        /api/v1
// @schemes         http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token dengan format: "Bearer <token>"
func main() {
	config.LoadEnv()

	dbConn := config.NewDB()
	
	defer dbConn.PgDB.Close()
	
	app := fiber.New() 

	route.RegisterAllRoutes(app, dbConn)

	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}