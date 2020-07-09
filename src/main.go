package main

import (
	"fmt"
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	InitSession()
	
	InitContents()

	var input string
	_, _ = fmt.Scanln(&input)
}

var (
	token string
	session *Session
)

func InitSession() {
	log.Println(token)

	if (len(token) == 0) {
		log.Fatal("Token can't be nil")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	session = &Session{
		token: token,
		bot: bot,
	}
	log.Println(`Session initialize successed`)

	session.SetHandler(func(s *Session, update tgbotapi.Update) {
		log.Println(update.Message.Text)

		context, err := GetContext(update.Message.Chat.ID)
		if err != nil {
			log.Println(err)
			return
		}

		if context == nil {
			context, err = CreateContext(update.Message.Chat.ID)
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "list": {
				response := context.HandleListCommand()
				session.Reply(context.id, update.Message.MessageID, response)
				break
			}

			case "subscribe": {
				args := update.Message.CommandArguments()
				response := context.HandleSubscribeCommand(args)
				session.Reply(context.id, update.Message.MessageID, response)
				break
			}

			case "unsubscribe": {
				args := update.Message.CommandArguments()
				response := context.HandleUnsubscribeCommand(args)
				session.Reply(context.id, update.Message.MessageID, response)
				break
			}

			default:
				break;
			}
		}
	})
	session.Run()
}

func InitContents() {
	err := LoadContexts()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(`Contents initialize successed`)
}