package main

import (
	"log"
	"fmt"
	"database/sql"
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

func (context *Context) Reload() {
	if context.id == 0 {
		return
	}

	cache := NewCache(context.id)
	context.cache = cache

	var monitors []*Monitor
	for _, subscription := range cache.GetSubscriptions() {
		monitor := NewMonitor(subscription)
		monitor.SetHandler(func(monitor *Monitor, items []*Item) (bool) {
			for _, item := range items {
				if cache.GetHasSent(item) {
					continue
				}

				msg := fmt.Sprintf("[%s](%s)", item.title, item.link)
				session.send(context.id, msg)

				cache.SetHasSent(item)
			}
			return false
		})
		monitor.Run()

		monitors = append(monitors, monitor)
	}

	context.monitors = monitors
}

func (context *Context) Save() (bool) {
	db, err := sql.Open("sqlite3", "context.db")
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists context (id int64 not null primary key);")
	if err != nil {
		log.Println(err)
		return false
	}
	
	_, err = db.Exec("replace into context(id) values(?);", context.id)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func NewContext(id int64) (context *Context) {
	return &Context {
		id: id,
	}
}

func GetContext(id int64) (*Context) {
	return contexts[id]
}

func LoadContexts() {
	db, err := sql.Open("sqlite3", "./contexts.db")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists context (id integer not null primary key);")
	if err != nil {
		log.Println(err)
		return
	}

	rows, err := db.Query("select * from contexts;")
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Println(err)
		}
		
		context := NewContext(id)
		context.Reload()
	}
}