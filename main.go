package main

import (
	"fmt"
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	session.SetHandler(func(s *Session, update tgbotapi.Update) (bool) {
		log.Println(update.Message.Text)

		context := GetContext(update.Message.Chat.ID)
		if context == nil {
			context = NewContext(update.Message.Chat.ID)
			context.Reload()
			context.Save()
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "list": {
				if len(context.monitors) == 0 {
					session.reply(context.id, update.Message.MessageID, "No subscription")
				} else {
					var message string
					for idx, monitor := range context.monitors {
						message += fmt.Sprintf("%d. [%s](%s) \n", idx + 1, monitor.subscription.title, monitor.subscription.link)
					}
					session.reply(context.id, update.Message.MessageID, message)
				}
				break
			}

			case "subscribe": {
				url := update.Message.CommandArguments()
				subscription, _ := Query(url)

				if subscription == nil {
					session.reply(context.id, update.Message.MessageID, "NO RSS Found")
				} else {
					context.cache.AddSubscription(subscription)
					context.Reload()

					session.reply(context.id, update.Message.MessageID, fmt.Sprintf(`Subscription [%s](%s) added.`, subscription.title, subscription.link))
				}
				break
			}

			default:
				break;
			}
		}
		return true
	})
	session.Run()

	LoadContexts()

	var input string
	_, _ = fmt.Scanln(&input)
}
