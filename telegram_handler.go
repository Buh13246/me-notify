package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
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
		} else if strings.HasPrefix(update.Message.Text, "/join") {
			args := strings.Fields(update.Message.Text)

			if len(args) < 2 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /join <token>")
				bot.Send(msg)
				continue
			}

			joinToken, err := uuid.Parse(args[1])

			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Token not valid")
				bot.Send(msg)
				continue
			}

			joinBotToken(bot, joinToken, update.Message.Chat.ID)

		} else if strings.HasPrefix(update.Message.Text, "/leave") {
			args := strings.Fields(update.Message.Text)

			if len(args) < 2 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /leave <token>")
				bot.Send(msg)
				continue
			}

			robot, err := uuid.Parse(args[1])

			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Token not valid")
				bot.Send(msg)
				continue
			}

			leaveBot(bot, robot, update.Message.Chat.ID)

		} else if strings.HasPrefix(update.Message.Text, "/robotcreate") {
			args := strings.Fields(update.Message.Text)
			fmt.Print(args)
			fmt.Printf("\n##########\n%d\n#########\n", len(args))
			if len(args) == 1 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /robotcreate <name of the new robot>")
				bot.Send(msg)
				continue
			}

			name := strings.Join(args[1:], " ")
			fmt.Printf("\nName: %s\n", name)

			token := uuid.New()
			joinToken := uuid.New()
			stmt, err := db.Prepare("INSERT INTO robots (Token, Owner, JoinToken, Name) VALUES (?, ?, ?, ?)")

			if err != nil {
				log.Fatalf("Could not Prepare Statement: %s\n", err)
			}

			_, err = stmt.Exec(token, update.Message.Chat.ID, joinToken, name)
			if err != nil {
				log.Panicf("Could not execute Statement: %s\n", err)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Created robot.\nToken: %s\nJoinToken: %s\nName: %s", token.String(), joinToken.String(), name))
			bot.Send(msg)
			joinBot(bot, token, update.Message.Chat.ID)
		} else if update.Message.Text == "/robots" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Test3")
			bot.Send(msg)
		} else if strings.HasPrefix(update.Message.Text, "/send") {
			args := strings.Fields(update.Message.Text)

			if len(args) < 3 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /send <token> <msg>")
				bot.Send(msg)
				continue
			}

			robot, err := uuid.Parse(args[1])

			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Token not valid")
				bot.Send(msg)
				continue
			}

			sendNotifyFromBot(bot, robot, strings.Join(args[2:], " "))

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
	rows.Close()

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

func sendNotifyFromBot(bot *tgbotapi.BotAPI, robot uuid.UUID, msgRaw string) {

	var robotName string

	stmt, err := db.Prepare("SELECT Name FROM robots WHERE Token = ?")

	if err != nil {
		log.Fatalf("Could not Prepare Statement: %s\n", err)
	}

	rows, err := stmt.Query(robot.String())
	defer rows.Close()

	if err != nil {
		log.Panicf("Could not query Statement: %s\n", err)
	}

	if !rows.Next() {
		return
	}

	rows.Scan(&robotName)

	stmt, err = db.Prepare("SELECT ChatID FROM subscribers WHERE Token = ?")

	if err != nil {
		log.Fatalf("Could not Prepare Statement: %s\n", err)
	}

	rows, err = stmt.Query(robot.String())
	defer rows.Close()
	if err != nil {
		log.Panicf("Could not query Statement: %s\n", err)
	}

	var chatID int64

	for {
		if !rows.Next() {
			break
		}

		rows.Scan(&chatID)
		fmt.Printf("######\nChatID: %d\n#######", chatID)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("<%s> %s", robotName, msgRaw))
		bot.Send(msg)
	}

}

func joinBotToken(bot *tgbotapi.BotAPI, joinToken uuid.UUID, chatID int64) {
	stmt, err := db.Prepare("SELECT Token, JoinToken FROM robots WHERE JoinToken = ?")

	if err != nil {
		log.Fatalf("Could not Prepare Statement: %s\n", err)
	}

	rows, err := stmt.Query(joinToken.String())
	defer rows.Close()

	if err != nil {
		log.Panicf("Could not execute Statement: %s\n", err)
	}

	if rows.Next() {
		var token string
		var jT string
		rows.Scan(&token, &jT)
		robot, err := uuid.Parse(token)

		fmt.Print(token)
		fmt.Print(jT)

		if err != nil {
			log.Fatalf("Could not parse UUID: %s\n", err)
			return
		}
		rows.Close()

		joinBot(bot, robot, chatID)
		return
	} else {
		msg := tgbotapi.NewMessage(chatID, "Token not valid")
		bot.Send(msg)
	}
}

func joinBot(bot *tgbotapi.BotAPI, robot uuid.UUID, chatID int64) {
	stmt, err := db.Prepare("SELECT * FROM subscribers WHERE Token = ? AND ChatID = ?")

	if err != nil {
		log.Fatalf("Could not Prepare Statement: %s\n", err)
	}

	rows, err := stmt.Query(robot.String(), chatID)
	defer rows.Close()

	if err != nil {
		log.Panicf("Could not execute Statement: %s\n", err)
	}

	if rows.Next() {
		msg := tgbotapi.NewMessage(chatID, "Already joined")
		bot.Send(msg)
		return
	}

	stmt, err = db.Prepare("INSERT INTO subscribers (Token, ChatID) VALUES (?, ?)")

	if err != nil {
		log.Fatalf("Could not Prepare Statement: %s\n", err)
	}

	_, err = stmt.Exec(robot.String(), chatID)
	if err != nil {
		log.Panicf("Could not execute Statement: %s\n", err)
	}

	msg := tgbotapi.NewMessage(chatID, "Successfully joined")
	bot.Send(msg)
}

func leaveBot(bot *tgbotapi.BotAPI, robot uuid.UUID, chatID int64) {

	stmt, err := db.Prepare("DELETE FROM subscribers WHERE Token = ? AND ChatID = ?")

	if err != nil {
		log.Fatalf("Could not Prepare Statement: %s\n", err)
	}

	_, err = stmt.Exec(robot.String(), chatID)
	if err != nil {
		log.Panicf("Could not execute Statement: %s\n", err)
	}

	msg := tgbotapi.NewMessage(chatID, "Successfully left")
	bot.Send(msg)
}
