package main

import (
	"math/rand"
	"time"
	"log"
	"fmt"
	"crypto/md5"
	"net/http"
	"github.com/mmcdole/gofeed"
)

type Monitor struct {
	subscription *Subscription
	handler func(monitor *Monitor, items []*Item) (bool)
}

func NewMonitor(subscription *Subscription) (*Monitor) {
	return &Monitor {
		subscription: subscription,
	}
}

func (monitor *Monitor) SetHandler(handler func(monitor *Monitor, items []*Item) (bool)) {
	monitor.handler = handler
}

func (monitor *Monitor) Run() {
	go monitor.Schedule(monitor.subscription)
}

func (monitor *Monitor) Schedule(subscription *Subscription) {
	monitor.Fetch()
	
	ticker := time.NewTicker(time.Duration(rand.Intn(config.MaxInterval - config.MinInterval) + config.MinInterval)*time.Second)
	stopTickers := make(chan bool)

	for {
		select {
		case <-stopTickers:
			return
		case <-ticker.C:
			monitor.Fetch()
		}
	}
}

func (monitor *Monitor) Fetch() {
	subscription, items := Query(monitor.subscription.link)
	if subscription == nil || items == nil || len(items) == 0 {
		return
	}

	monitor.handler(monitor, items)
}

func Query(url string) (*Subscription, []*Item) {
	req, err := http.NewRequest("GET", url, nil)
    if err != nil {
		return nil, nil
    }
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	subscription := &Subscription {
		id: fmt.Sprintf("%x", md5.Sum([]byte(feed.Link))),
		title: feed.Title,
		description: feed.Description,
		link: url,
	}

	var items []*Item
	for _, item := range feed.Items {
		items = append(items, &Item {
			id: item.GUID,
			title: item.Title,
			link: item.Link,
		})
	}

	return subscription, items
}