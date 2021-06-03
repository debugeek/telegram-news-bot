package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexflint/go-arg"
)

var args struct {
	Token string `arg:"-t,--token" help:"telegram bot token"`
}

func launch() {
	InitSession()

	InitMonitor()

	InitContents()
}

func main() {
	arg.MustParse(&args)

	token = args.Token
	if len(token) == 0 {
		log.Fatal("token not found")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go launch()

	<-sigs
}
