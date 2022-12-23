package main

import (
	"cw-cal/api"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

var (
	Poller = &tele.LongPoller{Timeout: 10 * time.Second}
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		panic("Error loading .env file")
	}
	pref := tele.Settings{
		Token:  os.Getenv("API_KEY"),
		Poller: Poller,
	}
	Bot := api.NewBot(pref)
	Bot.Start()
}
