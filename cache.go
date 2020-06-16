package main

import (
	"log"
	"fmt"
	"database/sql"
)

type Cache struct {
	id int64
}

func NewCache(id int64) (*Cache) {
	cache := &Cache {
		id: id,
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", id))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer db.Close()

	_, err = db.Exec(`
		create table if not exists subscription (id text not null primary key, title text, description text, link text);
		create table if not exists item (id text not null primary key);
	`)
	if err != nil {
		log.Println(err)
		return nil
	}

	return cache
}

func (cache *Cache) AddSubscription(subscription *Subscription) (bool) {
	if subscription == nil {
		return false
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	_, err = db.Exec("replace into subscription(id, title, description, link) values(?, ?, ?, ?);", subscription.id, subscription.title, subscription.description, subscription.link)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (cache *Cache) GetSubscriptions() ([]*Subscription) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer db.Close()

	rows, err := db.Query("select * from subscription;")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	var subscriptions []*Subscription

	for rows.Next() {
		var id string
		var title string
		var description string
		var link string
		if err := rows.Scan(&id, &title, &description, &link); err != nil {
			log.Println(err)
			continue
		}

		subscriptions = append(subscriptions, &Subscription {
			id: id,
			title: title,
			description: description,
			link: link,
		})
	}

	return subscriptions
}

func (cache *Cache) GetHasSent(item *Item) (bool) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	var exists bool
	row := db.QueryRow("select exists(select 1 from item where id = ?);", item.id)
	if err := row.Scan(&exists); err != nil {
    	return false
	}
	
	return exists
}

func (cache *Cache) SetHasSent(item *Item) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("replace into item(id) values(?);", item.id)
	if err != nil {
		log.Println(err)
		return
	}

	return
}