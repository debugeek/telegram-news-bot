package main

import (
	"log"
	"math/rand"
	"time"
)

type Monitor struct {
	subjects map[string][]func(items []*Item)
	ticker   *time.Ticker
	quit     chan bool
}

func InitMonitor() {
	SharedMonitor().Run()
	log.Println(`Monitor initialized`)
}

func SharedMonitor() *Monitor {
	monitorOnce.Do(func() {
		monitor = &Monitor{
			subjects: make(map[string][]func(items []*Item)),
		}
	})
	return monitor
}

func (monitor *Monitor) Observe(link string, handler func(items []*Item)) {
	handlers := monitor.subjects[link]
	if handlers == nil {
		handlers = make([]func(items []*Item), 0)
		monitor.subjects[link] = handlers
	}

	handlers = append(handlers, handler)
	monitor.subjects[link] = handlers

	monitor.Pull()
}

func (monitor *Monitor) Run() {
	go monitor.Launch()
}

func (monitor *Monitor) Stop() {
	monitor.quit <- true
}

func (monitor *Monitor) Launch() {
	monitor.Pull()

	monitor.ticker = time.NewTicker(time.Duration(rand.Intn(60)+300) * time.Second)
	monitor.quit = make(chan bool)

	for {
		select {
		case <-monitor.quit:
			return
		case <-monitor.ticker.C:
			monitor.Pull()
		}
	}
}

func (monitor *Monitor) Pull() {
	for link, handlers := range monitor.subjects {
		if len(handlers) == 0 {
			continue
		}

		items, err := FetchItems(link)
		if len(items) == 0 || err != nil {
			continue
		}

		for _, handler := range handlers {
			handler(items)
		}
	}

}
