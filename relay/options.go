package relay

type ServerOptions struct {
	Address string // defaults to ":8081"
}

func (o *ServerOptions) init() {
	if o.Address == "" {
		o.Address = ":8081"
	}
}
