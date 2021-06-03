package main

type Account struct {
	Id   int64 `firestore:"id"`
	kind int   `firestore:"kind"`
}

type Source struct {
	Id    string `firestore:"id"`
	Link  string `firestore:"link"`
	Title string `firestore:"title"`
}

type Subscription struct {
	Sources map[string]*Source `firestore:"sources"`
}

type Channel struct {
	id          string
	title       string
	description string
	link        string
}

type Item struct {
	guid  string
	title string
	link  string
}
