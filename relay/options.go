package relay

import (
	"github.com/moiji-mobile/redisrelay/relay/proto"
	"go.uber.org/zap"
	"time"
)

type ServerOptions struct {
	proto.ConfigProtoP
	Logger  *zap.Logger
	TimeOut time.Duration
}

func DefaultOptions() ServerOptions {
	logger, _ := zap.NewDevelopment()

	var defaultOptions = ServerOptions{
		Logger: logger}
	return defaultOptions
}
