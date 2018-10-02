package main

import (
	"fmt"
	"github.com/moiji-mobile/redisrelay/relay"
)

func main() {
	opt := relay.ServerOptions{
		Address: ":8080",
	}
	server, err := relay.NewServer(&opt)
	if err != nil {
		fmt.Println("Fooo")
	}
	server.Run()
}
