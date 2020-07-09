package main

import (
	"time"
)

type Subscription struct {
	id string
	title string
	description string
	link string
	date *time.Time
}

type Item struct {
	id string
	guid string
	title string
	link string
	date *time.Time
}