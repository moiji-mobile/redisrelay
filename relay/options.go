package relay

import (
	"go.uber.org/zap"
)

type ServerOptions struct {
	Address string // defaults to ":8081"
	Logger  *zap.Logger
}

func (o *ServerOptions) init() {
	if o.Address == "" {
		o.Address = ":8081"
	}

	if o.Logger == nil {
		o.Logger, _ = zap.NewDevelopment()
	}
}
