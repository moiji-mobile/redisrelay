package relay

import (
	"go.uber.org/zap"
)

type ServerOptions struct {
	BindAddress     string   `config:"bind_address"` // defaults to ":8081"
	RemoteAddresses []string `config:"remote_addresses"`
	Logger          *zap.Logger
}

func DefaultOptions() ServerOptions {
	logger, _ := zap.NewDevelopment()

	var defaultOptions = ServerOptions{
		BindAddress: ":8081",
		Logger:      logger,
	}
	return defaultOptions
}
