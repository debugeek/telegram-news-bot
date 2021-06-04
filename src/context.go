package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Context struct {
	id           int64
	account      *Account
	subscription *Subscription
}

func InitContents() error {
	accounts, err := SharedFirebase().GetAccounts()
	if err != nil {
		return err
	}

	for _, account := range accounts {
		_, err := NewContext(account.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewContext(id int64) (*Context, error) {
	context := contexts[id]
	if context != nil {
		return context, nil
	}

	account, err := SharedFirebase().GetAccount(id)
	if err != nil {
		return nil, err
	}
	if account == nil {
		account = &Account{
			Id:   id,
			kind: 0,
		}
		err = SharedFirebase().SaveAccount(account)
		if err != nil {
			return nil, err
		}
	}

	subscription, err := SharedFirebase().GetSubscription(account)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		subscription = &Subscription{
			Sources: make(map[string]*Source),
		}
		SharedFirebase().SaveSubscription(account, subscription)
	}

	context = &Context{
		id:           id,
		account:      account,
		subscription: subscription,
	}

	for _, source := range subscription.Sources {
		err = context.StartObserving(source)
		if err != nil {
			return nil, err
		}
		context.subscription.Sources[source.Id] = source
	}

	contexts[account.Id] = context

	return context, nil
}

func (context *Context) StartObserving(source *Source) error {
	observer := &Observer{
		identifier: context.id,
		handler: func(items []*Item) {
			if len(items) == 0 {
				return
			}

			records, err := SharedFirebase().GetPostRecords(context.account)
			if err != nil {
				log.Println(err)
				return
			}

			for _, item := range items {
				if records[item.guid] {
					continue
				}

				msg := fmt.Sprintf("[%s](%s)", item.title, item.link)
				err = session.Send(context.id, msg)
				if err != nil {
					log.Println(err)
					return
				}

				records[item.guid] = true
			}

			SharedFirebase().SetPostRecords(context.account, records)
		},
	}
	SharedMonitor().AddObserver(observer, source.Link)

	return nil
}

func (context *Context) StopObserving(source *Source) error {
	SharedMonitor().RemoveObserver(context.id, source.Link)

	return nil
}

func (context *Context) Subscribe(channel *Channel) (*Source, error) {
	id := channel.id

	source := context.subscription.Sources[id]
	if source != nil {
		return nil, fmt.Errorf(`Source [%s](%s) exists`, source.Title, source.Link)
	}

	source = &Source{
		Id:        id,
		Link:      channel.link,
		Title:     channel.title,
		Timestamp: time.Now().Unix(),
	}
	context.subscription.Sources[id] = source

	err := fb.SaveSubscription(context.account, context.subscription)
	if err != nil {
		return nil, err
	}

	return source, nil
}

func (context *Context) Unsubscribe(source *Source) error {
	delete(context.subscription.Sources, source.Id)

	err := fb.SaveSubscription(context.account, context.subscription)

	return err
}

func (context *Context) MarkItemsPosted(items []*Item) error {
	records, err := SharedFirebase().GetPostRecords(context.account)
	if err != nil {
		return err
	}

	for _, item := range items {
		records[item.guid] = true
	}

	return SharedFirebase().SetPostRecords(context.account, records)
}

func (context *Context) GetSources() []*Source {
	sources := make([]*Source, 0)
	for _, source := range context.subscription.Sources {
		sources = append(sources, source)
	}

	sort.SliceStable(sources, func(i, j int) bool {
		return sources[i].Timestamp < sources[j].Timestamp
	})

	return sources
}

// Handlers

func (context *Context) HandleListCommand() string {
	sources := context.GetSources()
	if len(sources) == 0 {
		return `No source found`
	}

	var message string
	for idx, source := range sources {
		message += fmt.Sprintf("%d. [%s](%s) \n", idx+1, source.Title, source.Link)
	}
	return message
}

func (context *Context) HandleSubscribeCommand(args string) string {
	if len(args) == 0 || !isValidURL(args) {
		return `Please input a valid url.`
	}

	if channel, items, err := FetchChannel(args); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if source, err := context.Subscribe(channel); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.MarkItemsPosted(items); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.StartObserving(source); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else {
		if len(items) == 0 {
			return fmt.Sprintf(`Channel [%s](%s) added.`, channel.title, channel.link)
		} else {
			return fmt.Sprintf(`Channel [%s](%s) added.
			
[%s](%s)`, channel.title, channel.link, items[0].title, items[0].link)
		}
	}
}

func (context *Context) HandleUnsubscribeCommand(args string) string {
	sources := context.GetSources()

	index, err := strconv.Atoi(args)
	if err != nil || index <= 0 || index > len(sources) {
		return `Please input a valid index.`
	}

	index -= 1

	source := sources[index]

	if err := context.Unsubscribe(source); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.StopObserving(source); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else {
		return fmt.Sprintf(`Subscrption %s deleted.`, source.Title)
	}
}

func (context *Context) HandleStatisticCommand(args string) string {
	sources := context.GetSources()

	index, err := strconv.Atoi(args)
	if err != nil || index <= 0 || index > len(sources) {
		return `Please input a valid index.`
	}

	index -= 1

	source := sources[index]

	if err := context.Unsubscribe(source); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.StopObserving(source); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else {
		return fmt.Sprintf(`Subscrption %s deleted.`, source.Title)
	}
}
