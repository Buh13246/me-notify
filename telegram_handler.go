package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Function that continuously listens for new users.
func registerNewUsers(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.Text == "/register" {
			go userRegister(bot, update.Message.Chat.ID)
		} else if update.Message.Text == "/start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please use /register to register")
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
			bot.Send(msg)
		}

	}
}

func userRegister(bot *tgbotapi.BotAPI, chatID int64) {
	stmt, err := db.Prepare("SELECT * FROM users WHERE ChatID = ?")
	rows, err := stmt.Query(chatID)
	defer rows.Close()

	if err != nil {
		log.Fatalf("Could not execute SELECT: %s\n", err)
	}

	if rows.Next() {
		msg := tgbotapi.NewMessage(chatID, "Already registered")
		bot.Send(msg)
		return
	}

	stmt, err = db.Prepare("INSERT INTO users( ChatID ) VALUES (?);")

	if err != nil {
		log.Fatalf("Could not create user: %s\n", err)
	}

	_, err = stmt.Exec(chatID)
	if err != nil {
		log.Fatalf("Failed to exec statement: %s\n", err)
	}

	msg := tgbotapi.NewMessage(chatID, "Welcome to the resource notifier bot! You have been registered to receive alerts.")
	bot.Send(msg)
}
