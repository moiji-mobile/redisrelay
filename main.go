package main

import (
	"flag"
	"fmt"
	"github.com/elastic/go-ucfg"
	"github.com/elastic/go-ucfg/yaml"
	"github.com/moiji-mobile/redisrelay/relay"
	"os"
)

var (
	configPath = flag.String("config_path", "config.yml", "The path to the config file")
)

func main() {
	// Read the flags
	flag.Parse()

	// Parse the options
	appOpts := relay.DefaultOptions()
	config, err := yaml.NewConfigWithFile(*configPath, ucfg.PathSep("."))
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	err = config.Unpack(&appOpts)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// Create the server and start
	fmt.Printf("Using config: %#v\n", appOpts)
	server, err := relay.NewServer(&appOpts)
	if err != nil {
		fmt.Println("Fooo")
	}
	server.Run()
}
