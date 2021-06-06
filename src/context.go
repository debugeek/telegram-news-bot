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
	id            int64
	account       *Account
	subscriptions map[string]*Subscription
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

	subscriptions, err := SharedFirebase().GetSubscriptions(account)
	if err != nil {
		return nil, err
	}
	if subscriptions == nil {
		subscriptions = make(map[string]*Subscription)
	}

	context = &Context{
		id:            id,
		account:       account,
		subscriptions: subscriptions,
	}

	for _, subscription := range subscriptions {
		err = context.StartObserving(subscription)
		if err != nil {
			return nil, err
		}
	}

	contexts[account.Id] = context

	return context, nil
}

func (context *Context) StartObserving(subscription *Subscription) error {
	observer := &Observer{
		identifier: context.id,
		handler: func(items []*Item) {
			if len(items) == 0 {
				return
			}

			itemIds := make([]string, 0)

			for _, item := range items {
				pushed, err := SharedFirebase().GetItemPushed(context.account, item.id)
				if pushed {
					continue
				}
				if err != nil {
					log.Println(err)
					continue
				}

				msg := fmt.Sprintf("[%s](%s)", item.title, item.link)
				err = session.Send(context.id, msg)
				if err != nil {
					log.Println(err)
					return
				}

				itemIds = append(itemIds, item.id)
			}

			SharedFirebase().SetItemsPushed(context.account, itemIds)
		},
	}
	SharedMonitor().AddObserver(observer, subscription.Link)

	return nil
}

func (context *Context) StopObserving(subscription *Subscription) error {
	SharedMonitor().RemoveObserver(context.id, subscription.Link)

	return nil
}

func (context *Context) Subscribe(channel *Channel) (*Subscription, error) {
	id := channel.id

	subscription := context.subscriptions[id]
	if subscription != nil {
		return nil, fmt.Errorf(`Subscription [%s](%s) exists`, subscription.Title, subscription.Link)
	}

	subscription = &Subscription{
		Id:        id,
		Link:      channel.link,
		Title:     channel.title,
		Timestamp: time.Now().Unix(),
	}
	context.subscriptions[id] = subscription

	err := fb.AddSubscription(context.account, subscription)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

func (context *Context) Unsubscribe(subscription *Subscription) error {
	delete(context.subscriptions, subscription.Id)

	err := fb.DeleteSubscription(context.account, subscription)

	return err
}

func (context *Context) SetItemsPushed(items []*Item) error {
	itemIds := make([]string, 0)
	for _, item := range items {
		itemIds = append(itemIds, item.id)
	}
	return SharedFirebase().SetItemsPushed(context.account, itemIds)
}

func (context *Context) GetSubscriptions() []*Subscription {
	subscriptions := make([]*Subscription, 0)
	for _, subscription := range context.subscriptions {
		subscriptions = append(subscriptions, subscription)
	}

	sort.SliceStable(subscriptions, func(i, j int) bool {
		return subscriptions[i].Timestamp < subscriptions[j].Timestamp
	})

	return subscriptions
}

// Handlers

func (context *Context) HandleListCommand() string {
	subscriptions := context.GetSubscriptions()
	if len(subscriptions) == 0 {
		return `No subscription found`
	}

	var message string
	for idx, subscription := range subscriptions {
		message += fmt.Sprintf("%d. [%s](%s) \n", idx+1, subscription.Title, subscription.Link)
	}
	return message
}

func (context *Context) HandleSubscribeCommand(args string) string {
	if len(args) == 0 || !isValidURL(args) {
		return `Please input a valid url.`
	}

	if channel, items, err := FetchChannel(args); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if subscription, err := context.Subscribe(channel); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.SetItemsPushed(items); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.StartObserving(subscription); err != nil {
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
	subscriptions := context.GetSubscriptions()

	index, err := strconv.Atoi(args)
	if err != nil || index <= 0 || index > len(subscriptions) {
		return `Please input a valid index.`
	}

	index -= 1

	subscription := subscriptions[index]

	if err := context.Unsubscribe(subscription); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.StopObserving(subscription); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else {
		return fmt.Sprintf(`Subscrption %s deleted.`, subscription.Title)
	}
}

func (context *Context) HandleStatisticCommand(args string) string {
	if statistics, err := SharedFirebase().GetTopSubscriptions(5); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if len(statistics) == 0 {
		return "Insufficient data."
	} else {
		var message string
		for idx, statistic := range statistics {
			message += fmt.Sprintf("%d. [%s](%s) (%d)\n", idx+1, statistic.Subscription.Title, statistic.Subscription.Link, statistic.Count)
		}
		return message
	}

}
