package main

import (
	"fmt"
	"database/sql"
	"strconv"
	_ "github.com/mattn/go-sqlite3"
)

type Context struct {
	id int64
	cache *Cache
	monitors []*Monitor
}

var (
	contexts map[int64]*Context
)

func init() {
	contexts = make(map[int64]*Context)
}

func NewContext(id int64) (context *Context) {
	return &Context {
		id: id,
		cache: NewCache(id),
	}
}

func GetContext(id int64) (*Context) {
	return contexts[id]
}

func LoadContexts() (error) {
	db, err := sql.Open("sqlite3", "./context.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists context (id integer not null primary key);")
	if err != nil {
		return err
	}

	rows, err := db.Query("select * from context;")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		
		context := NewContext(id)

		subscriptions, err := context.cache.GetSubscriptions()
		if err != nil {
			return err
		}

		for _, subscription := range subscriptions {
			context.Enqueue(subscription)
		}

		contexts[id] = context
	}

	return nil
}


func (context *Context) Create() (error) {
	err := context.cache.Create()
	if err != nil {
		return err
	}
	return nil
}

func (context *Context) Enqueue(subscription *Subscription) (error) {
	exists := false
	for _, monitor := range context.monitors {
		if monitor.subscription.id == subscription.id {
			exists = true
			break
		}
	}

	if exists {
		return fmt.Errorf(`Subscription [%s](%s) exists.`, subscription.title, subscription.link)
	}

	err := context.cache.AddSubscription(subscription)
	if err != nil {
		return err
	}

	monitor := NewMonitor(subscription)
	monitor.SetHandler(func(monitor *Monitor, items []*Item) {
		for _, item := range items {
			if hasSent, err := context.cache.GetHasSent(item); hasSent || err != nil {
				continue
			}

			msg := fmt.Sprintf("[%s](%s)", item.title, item.link)
			session.Send(context.id, msg)

			context.cache.SetHasSent(item)
		}
	})
	monitor.Run()

	context.monitors = append(context.monitors, monitor)

	return nil
}

func (context *Context) Dequeue(subscription *Subscription) (error) {
	err := context.cache.DeleteSubscription(subscription)
	if err != nil {
		return err
	}

	for index, monitor := range context.monitors {
		if monitor.subscription.id != subscription.id {
			continue
		}

		monitor.Stop()

		context.monitors = append(context.monitors[:index], context.monitors[index + 1:]...)
	}

	return nil;
}

func (context *Context) Save() (error) {
	db, err := sql.Open("sqlite3", "context.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists context (id int64 not null primary key);")
	if err != nil {
		return err
	}
	
	_, err = db.Exec("replace into context(id) values(?);", context.id)
	if err != nil {
		return err
	}

	contexts[context.id] = context

	return nil
}



// Handlers

func (context *Context) HandleListCommand() (string) {
	if len(context.monitors) == 0 {
		return "No currently subscription now."
	} else {
		var message string
		for idx, monitor := range context.monitors {
			message += fmt.Sprintf("%d. [%s](%s) \n", idx + 1, monitor.subscription.title, monitor.subscription.link)
		}
		return message
	}
}

func (context *Context) HandleSubscribeCommand(args string) (string) {
	if len(args) == 0 || !isValidURL(args) {
		return `Please input a valid url.`
	}
	
	
	if subscription, _, err := Query(args); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.Enqueue(subscription); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else {
		return fmt.Sprintf(`Subscription [%s](%s) added.`, subscription.title, subscription.link)
	}
}

func (context *Context) HandleUnsubscribeCommand(args string) (string) {
	index, err := strconv.Atoi(args)
	if err != nil || index <= 0 || index > len(context.monitors) {
		return `Please input a valid index.`
	}

	index -= 1

	subscription := context.monitors[index].subscription

	if err := context.Dequeue(subscription); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else {
		return fmt.Sprintf(`Subscrption %s deleted.`, subscription.title)
	}
}
