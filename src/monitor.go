package main

import (
	"math/rand"
	"time"
)

type Monitor struct {
	subscription *Subscription
	handler func(monitor *Monitor, items []*Item)

	ticker *time.Ticker
	quit chan bool
}

func (monitor *Monitor) SetHandler(handler func(monitor *Monitor, items []*Item)) {
	monitor.handler = handler
}

func (monitor *Monitor) Run() {
	go monitor.Schedule(monitor.subscription)
}

func (monitor *Monitor) Stop() {
	monitor.quit <- true
}

func (monitor *Monitor) Schedule(subscription *Subscription) {
	monitor.Refresh()
	
	monitor.ticker = time.NewTicker(time.Duration(rand.Intn(60) + 300)*time.Second)
	monitor.quit = make(chan bool)

	for {
		select {
		case <-monitor.quit:
			return
		case <-monitor.ticker.C:
			monitor.Refresh()
		}
	}
}

func (monitor *Monitor) Refresh() {
	items, err := FetchItems(monitor.subscription.link)
	if items == nil || len(items) == 0 || err != nil {
		return
	}
	
	for _, item := range items {
		item.id = monitor.subscription.id
	}

	monitor.handler(monitor, items)
}

