package main

import (
	"time"
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Session struct {
	bot *tgbotapi.BotAPI
	token string
	handler func(s *Session, update tgbotapi.Update)
}

func (session *Session) SetHandler(handler func(s *Session, update tgbotapi.Update)) {
	session.handler = handler
}

func (session *Session) Run() {
	go session.Schedule()
}

func (session *Session) Schedule() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updates, err := session.bot.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
		return
	}

	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		session.handler(session, update)
	}
}

func (session *Session) Send(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "markdown"
	session.bot.Send(msg)
}

func (session *Session) Reply(chatID int64, replyToMessageID int, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "markdown"
	msg.ReplyToMessageID = replyToMessageID
	session.bot.Send(msg)
}