package relay

import (
	"fmt"
	"go.uber.org/zap"
	"net"
)

type Server struct {
	listener net.Listener
	logger   *zap.Logger
	options  *ServerOptions
}

// A pair of reader/writer for a downstream connection
type downStream struct {
	conn   net.Conn
	reader *Reader
	writer *Writer
}

type remote struct {
	network string
	address string
	logger  *zap.Logger

	// Internal stream handling
	streams []*downStream
}

// A single client.
type Client struct {
	server  *Server
	reader  *Reader
	writer  *Writer
	remotes []remote
}

// A result coming from downstream connection
type forwardResult struct {
	result interface{}
	err    error
}

func NewServer(opt *ServerOptions) (*Server, error) {
	if len(opt.RemoteAddresses) == 0 {
		return nil, fmt.Errorf("Need to have at least one downstream: %v", len(opt.RemoteAddresses))
	}

	ln, err := net.Listen("tcp", opt.BindAddress)
	if err != nil {
		return nil, err
	}

	s := Server{listener: ln}
	s.options = opt
	s.logger = opt.Logger
	return &s, nil
}

// First thing that might be connection pooling..
func (remote *remote) getDownStream() (*downStream, error) {
	remoteConn, err := net.Dial(remote.network, remote.address)
	if err != nil {
		remote.logger.Error("Can't connect to real Redis", zap.Error(err))
		return nil, err
	}

	stream := &downStream{
		conn:   remoteConn,
		reader: NewReader(remoteConn, remote.logger),
		writer: NewWriter(remoteConn)}

	// Check the connection is working
	res, err := stream.sendReceive([]interface{}{[]byte{'p', 'i', 'n', 'g'}, []byte{'1', '2', '3'}}, remote.logger)
	if err != nil {
		remote.logger.Error("Can't ping", zap.Error(err))
		return nil, err
	}
	if string(*res.(*[]byte)) != "123" {
		return nil, fmt.Errorf("Sequencing error got: '%v'", res)
	}
	return stream, nil
}

// Return it to the pool
func (remote *remote) releaseDownStream(stream *downStream) {
	// Just close it now as streams are not tracked..
	stream.conn.Close()
}

func (remote *remote) forwardCommand(c chan<- forwardResult, cmd interface{}, logger *zap.Logger) {
	stream, err := remote.getDownStream()
	if err != nil {
		c <- forwardResult{result: nil, err: err}
	} else {
		res, err := stream.sendReceive(cmd, logger)
		c <- forwardResult{result: res, err: err}
	}
	remote.releaseDownStream(stream)
}

func (stream *downStream) sendReceive(cmd interface{}, logger *zap.Logger) (interface{}, error) {
	// Write it down stream
	err := stream.writer.Write(cmd)
	if err != nil {
		logger.Error("Can't forward command", zap.Error(err))
		return nil, err
	}
	stream.writer.Flush()

	// Get the response
	res, err := stream.reader.ParseData()
	if err != nil {
		logger.Error("Can't read response", zap.Error(err))
		return nil, err
	}
	return res, err
}

func selectResult(results []forwardResult, errors []forwardResult) (interface{}, error) {
	// Pick any success and then any error.
	if len(results) > 0 {
		return results[0].result, results[0].err
	}
	if len(errors) > 0 {
		return errors[0].result, errors[0].err
	}
	return nil, fmt.Errorf("Neither success nor failure. Reporting as failed")
}

func forwardDownstream(client *Client, cmd interface{}, logger *zap.Logger) (interface{}, error) {
	c := make(chan forwardResult, len(client.remotes))
	failures := make([]forwardResult, 0, len(client.remotes))
	results := make([]forwardResult, 0, len(client.remotes))

	for _, remote := range client.remotes {
		go remote.forwardCommand(c, cmd, logger)
	}

	for _, _ = range client.remotes {
		f := <-c
		if f.err != nil {
			failures = append(failures, f)
		} else {
			results = append(results, f)
		}
	}
	return selectResult(results, failures)
}

func (client *Client) forwardCommands() {
	for {
		// Get the command
		cmd, err := client.reader.ParseCommand()
		if err != nil {
			client.server.logger.Error("Can't parse command", zap.Error(err))
			return
		}

		res, err := forwardDownstream(client, cmd, client.server.logger)
		if err != nil {
			client.server.logger.Error("Can't parse command", zap.Error(err))
			return
		}

		// Write the response
		err = client.writer.Write(res)
		if err != nil {
			client.server.logger.Error("Can't forward response", zap.Error(err))
			return
		}
		client.writer.Flush()
	}
}

func handleConnection(conn net.Conn, s *Server) {
	defer conn.Close()

	client := Client{
		server: s,
		reader: NewReader(conn, s.logger),
		writer: NewWriter(conn)}

	client.remotes = make([]remote, 0)
	for _, addr := range s.options.RemoteAddresses {
		r := remote{
			logger:  s.logger,
			network: "tcp",
			address: addr}
		client.remotes = append(client.remotes, r)
	}
	client.forwardCommands()
}

func (s *Server) Run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error("Can't accept connection", zap.Error(err))
			continue
		}
		// Start a new go-routine to handle this.
		go handleConnection(conn, s)
	}
}
