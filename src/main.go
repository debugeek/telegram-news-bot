package main

import (
	"fmt"
	"log"

	"github.com/alexflint/go-arg"
)

var args struct {
	Token string `arg:"-t,--token" help:"telegram bot token"`
}

func main() {
	arg.MustParse(&args)

	token = args.Token
	if len(token) == 0 {
		log.Fatal("token not found")
	}

	InitSession()

	InitMonitor()

	InitContents()

	var input string
	_, _ = fmt.Scanln(&input)
}
