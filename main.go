package main

import (
	"flag"
	"fmt"
	"github.com/moiji-mobile/redisrelay/relay"
	//"github.com/moiji-mobile/redisrelay/relay/proto"
	"github.com/golang/protobuf/jsonpb"
	"os"
)

var (
	configPath = flag.String("config_path", "config.json", "The path to the config file")
)

func main() {
	// Read the flags
	flag.Parse()

	// Parse the options
	appOpts := relay.DefaultOptions()
	//appOpts := proto.ConfigProtoP{}

	r, err := os.Open(*configPath)
	if err != nil {
		fmt.Printf("Failed to open: %v\n", err)
		os.Exit(1)
	}
	err = jsonpb.Unmarshal(r, &appOpts.ConfigProtoP)
	if err != nil {
		fmt.Printf("Failed to parse: %v\n", err)
		os.Exit(1)
	}

	// Create the server and start
	fmt.Printf("Using config: %#v\n", appOpts)
	server, err := relay.NewServer(&appOpts)
	if err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		return
	}
	server.Run()
}
