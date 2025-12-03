package models

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var database *gorm.DB
	var err error

	dbURL := os.Getenv("DB_URL") // Render gives us this

	if dbURL == "" {
		// Use SQLite locally if no DB_URL is found
		fmt.Println("Using Local SQLite DB...")
		database, err = gorm.Open(sqlite.Open("expenses.db"), &gorm.Config{})
	} else {
		// Use Postgres if DB_URL exists
		fmt.Println("Using PostgreSQL DB...")
		database, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	}

	if err != nil {
		// This will print the EXACT reason why it failed in the logs
		panic("Failed to connect to database: " + err.Error())
	}

	database.AutoMigrate(&User{}, &Expense{})
	DB = database
}
