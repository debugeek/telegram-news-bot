package main

type Config struct {
	Token string `json:"token"`
	MinInterval int `json:"minimum_interval"`
	MaxInterval int `json:"maximum_interval"`
}

type Subscription struct {
	id string
	title string
	description string
	link string
}

type Item struct {
	id string
	guid string
	title string
	link string
}