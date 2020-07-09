package main

import (
	"log"
	"fmt"
	"time"
	"database/sql"
)

type Database struct {
	id int64
}

func CreateDatabase(id int64) (*Database, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", id))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec(`create table if not exists subscription (id text not null primary key, title text, description text, link text, latest_publish_date datetime);`)
	if err != nil {
		return nil, err
	}

	database := &Database {
		id: id,
	}
	return database, nil
}

func (database *Database) QuerySubscriptions() ([]*Subscription, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", database.id))
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
		var date *time.Time
		if err := rows.Scan(&id, &title, &description, &link, &date); err != nil {
			log.Println(err)
			continue
		}

		subscriptions = append(subscriptions, &Subscription {
			id: id,
			title: title,
			description: description,
			link: link,
			date: date,
		})
	}

	return subscriptions, nil
}

func (database *Database) InsertSubscription(subscription *Subscription) (error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", database.id))
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

func (database *Database) UpdateSubscription(subscription *Subscription) (error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", database.id))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("update subscription set title = ?, description = ?, link = ?, latest_publish_date = ? where id = ?;", subscription.title, subscription.description, subscription.link, subscription.date, subscription.id)
	if err != nil {
		return err
	}
	
	return nil
}

func (database *Database) DeleteSubscription(subscription *Subscription) (error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%d.db", database.id))
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