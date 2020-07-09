package main

import (
	"fmt"
	"database/sql"
	"strconv"
	_ "github.com/mattn/go-sqlite3"
)

type Context struct {
	id int64
	database *Database
	monitors []*Monitor
}

var contexts map[int64]*Context = make(map[int64]*Context)

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
		
		context, err := GetContext(id)
		if err != nil {
			return err
		}

		subscriptions, err := context.database.QuerySubscriptions()
		if err != nil {
			return err
		}

		for _, subscription := range subscriptions {
			context.EnqueueSubscription(subscription)
		}

		contexts[id] = context
	}

	return nil
}

func CreateContext(id int64) (*Context, error) {
	db, err := sql.Open("sqlite3", "context.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists context (id int64 not null primary key);")
	if err != nil {
		return nil, err
	}
	
	_, err = db.Exec("replace into context(id) values(?);", id)
	if err != nil {
		return nil, err
	}

	database, err := CreateDatabase(id)
	if err != nil {
		return nil, err
	}

	context := &Context {
		id: id,
		database: database,
		monitors: make([]*Monitor, 0),
	}

	contexts[id] = context

	return context, nil
}

func GetContext(id int64) (*Context, error) {
	context := contexts[id]
	if context != nil {
		return context, nil
	}

	db, err := sql.Open("sqlite3", "context.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select * from context where id = ?;", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		database := &Database {
			id: id,
		}
		context = &Context {
			id: id,
			database: database,
		}
		contexts[id] = context
	}

	return context, nil
}

func (context *Context) SubscribeSubscription(subscription *Subscription) (error) {
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

	err := context.database.InsertSubscription(subscription)
	if err != nil {
		return err
	}

	return nil
}

func (context *Context) UnsubscribeSubscription(subscription *Subscription) (error) {
	err := context.database.DeleteSubscription(subscription)
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

func (context *Context) EnqueueSubscription(subscription *Subscription) (error) {
	for _, monitor := range context.monitors {
		if monitor.subscription.id == subscription.id {
			return fmt.Errorf(`Subscription [%s](%s) duplicated.`, subscription.title, subscription.link)
		}
	}

	monitor := &Monitor {
		subscription: subscription,
	}

	monitor.SetHandler(func(monitor *Monitor, items []*Item) {
		if len(items) == 0 {
			return
		}

		subscription := monitor.subscription

		if subscription.date != nil {
			for _, item := range items {
				if item.date.After(*subscription.date) {
					msg := fmt.Sprintf("[%s](%s)", item.title, item.link)
					session.Send(context.id, msg)
				}
			}
		}

		subscription.date = items[len(items) - 1].date
		context.database.UpdateSubscription(subscription)
	})
	monitor.Run()

	context.monitors = append(context.monitors, monitor)

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
	
	if subscription, items, err := FetchSubscription(args); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.SubscribeSubscription(subscription); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else if err := context.EnqueueSubscription(subscription); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else {
		if len(items) == 0 {
			return fmt.Sprintf(`Subscription [%s](%s) added.`, subscription.title, subscription.link)
		} else {
			return fmt.Sprintf(`Subscription [%s](%s) added.
			
[%s](%s)`, subscription.title, subscription.link, items[0].title, items[0].link)
		}
	}
}

func (context *Context) HandleUnsubscribeCommand(args string) (string) {
	index, err := strconv.Atoi(args)
	if err != nil || index <= 0 || index > len(context.monitors) {
		return `Please input a valid index.`
	}

	index -= 1

	subscription := context.monitors[index].subscription

	if err := context.UnsubscribeSubscription(subscription); err != nil {
		return fmt.Sprintf(`%s`, err)
	} else {
		return fmt.Sprintf(`Subscrption %s deleted.`, subscription.title)
	}
}