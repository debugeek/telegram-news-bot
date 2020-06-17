package main

import (
	"fmt"
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	session.SetHandler(func(s *Session, update tgbotapi.Update) {
		log.Println(update.Message.Text)

		context := GetContext(update.Message.Chat.ID)
		if context == nil {
			context = NewContext(update.Message.Chat.ID)
			context.Create()
			context.Save()
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

	err := LoadContexts()
	if err != nil {
		log.Println(err)
	}

	var input string
	_, _ = fmt.Scanln(&input)
}



