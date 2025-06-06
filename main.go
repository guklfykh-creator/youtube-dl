package main

import (
	"log"
	"time"

	tb "gopkg.in/telebot.v3"
)

var bot *tb.Bot

func main() {
	cfg := LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config error: %v", err)
	}

	pref := tb.Settings{
		Token:  cfg.BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatalf("bot init failed: %v", err)
	}
	bot = b

	registerHandlers()

	log.Println("bot started successfully")
	b.Start()
}

func registerHandlers() {
	bot.Handle("/start", onStart)
	bot.Handle("/help", onHelp)
	bot.Handle("/cancel", onCancel)
	bot.Handle(tb.OnText, onText)

	for _, btn := range qualityButtons {
		bot.Handle(btn, onQualitySelect)
	}
}