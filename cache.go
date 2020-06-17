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
	return &Cache {
		id: id,
	}
}

func (cache *Cache) Create() (error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		create table if not exists subscription (id text not null primary key, title text, description text, link text);
		create table if not exists item (id text not null, guid text not null);
	`)
	if err != nil {
		return err
	}

	return nil
}

func (cache *Cache) Delete() (error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		delete from subscription where id = ?;
		delete from item where id = ?;
	`, cache.id, cache.id)
	if err != nil {
		return err
	}

	return nil
}

func (cache *Cache) AddSubscription(subscription *Subscription) (error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("replace into subscription(id, title, description, link) values(?, ?, ?, ?);", subscription.id, subscription.title, subscription.description, subscription.link)
	if err != nil {
		return err
	}

	return nil
}

func (cache *Cache) GetSubscriptions() ([]*Subscription, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select * from subscription;")
	if err != nil {
		return nil, err
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

	return subscriptions, nil
}

func (cache *Cache) DeleteSubscription(subscription *Subscription) (error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("delete from subscription where id = ?;", subscription.id)
	if err != nil {
		return err
	}

	return nil
}

func (cache *Cache) GetHasSent(item *Item) (bool, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		return false, err
	}
	defer db.Close()

	var exists bool
	row := db.QueryRow("select exists(select 1 from item where id = ? and guid = ?);", item.id, item.guid)
	if err := row.Scan(&exists); err != nil {
    	return false, err
	}
	
	return exists, nil
}

func (cache *Cache) SetHasSent(item *Item) (error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", cache.id))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("insert into item(id, guid) values(?, ?);", item.id, item.guid)
	if err != nil {
		return err
	}

	return nil
}