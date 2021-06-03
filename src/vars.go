package main

import "sync"

var (
	token string

	sessionOnce sync.Once
	session     *Session

	firebaseOnce sync.Once
	fb           Firebase

	monitorOnce sync.Once
	monitor     *Monitor

	contexts map[int64]*Context = make(map[int64]*Context)
)
