package main

import (
	"fmt"
	"log"
	"sort"
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
