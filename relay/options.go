package relay

import (
	"go.uber.org/zap"
	"time"
)

type ServerOptions struct {
	BindAddress     string   `config:"bind_address"` // defaults to ":8081"
	RemoteAddresses []string `config:"remote_addresses"`
	TimeOut_base	string   `config:"request_timeout"`
	TimeOut		time.Duration
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
