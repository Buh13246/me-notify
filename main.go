package main

import (
	"log"
	"os"

	"database/sql"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

// List of chat IDs registered to receive alerts.
var userList []string
var db *sql.DB

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3",
		"db.sqlite")
	if err != nil {
		log.Fatalf("Can't access DB... Error: %s\n", err)
	}

	return db
}

func addDatabaseSchema() {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users( ChatID int(64) PRIMARY KEY NOT NULL);")

	if err != nil {
		log.Fatalf("Could not create Scheme\nError: %s\n", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS robots( Token VARCHAR(36) PRIMARY KEY NOT NULL, Owner int(64) NOT NULL, JoinToken VARCHAR(36) NOT NULL, Name VARCHAR (255) NOT NULL );")

	if err != nil {
		log.Fatalf("Could not create Scheme\nError: %s\n", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS subscribers( Token VARCHAR(36) NOT NULL, ChatID int(64) NOT NULL);")

	if err != nil {
		log.Fatalf("Could not create Scheme\nError: %s\n", err)
	}

}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	db = initDB()
	defer db.Close()
	addDatabaseSchema()
	registerNewUsers(bot)
}
