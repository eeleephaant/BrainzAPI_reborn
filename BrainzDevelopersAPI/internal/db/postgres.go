package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var AuthDB *gorm.DB
var DevsDB *gorm.DB

func Connect() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbnameAuth := os.Getenv("DB_AUTH_NAME")
	dbnameDevs := os.Getenv("DB_DEVS_NAME")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbnameAuth, port,
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	AuthDB = database

	dsn1 := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbnameDevs, port,
	)

	database1, err1 := gorm.Open(postgres.Open(dsn1), &gorm.Config{})
	if err1 != nil {
		log.Fatal("failed to connect database:", err)
	}
	DevsDB = database1
}
