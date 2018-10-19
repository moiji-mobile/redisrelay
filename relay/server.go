package relay

import (
	"fmt"
	"net"
	"go.uber.org/zap"
)

type Server struct {
	listener net.Listener
	logger *zap.Logger
}

// A pair of reader/writer for a downstream connection
type DownStream struct {
	reader *Reader
	writer *Writer
}

// A single client.
type Client struct {
	server *Server
	reader *Reader
	writer *Writer
	streams []DownStream
}

func NewServer(opt *ServerOptions) (*Server, error) {
	opt.init()

	ln, err := net.Listen("tcp", opt.Address)
	if err != nil {
		return nil, err
	}

	s := Server{listener: ln}
	s.logger = opt.Logger
	return &s, nil
}

func (remote *DownStream) sendReceive(cmd interface{}, logger *zap.Logger) (interface{}, error) {
	// Write it down stream
	err := remote.writer.Write(cmd)
	if err != nil {
		logger.Error("Can't forward command", zap.Error(err))
		return nil, err
	}
	remote.writer.Flush()

	// Get the response
	res, err := remote.reader.ParseData()
	if err != nil {
		logger.Error("Can't read response", zap.Error(err))
		return nil, err
	}
	return res, err
}

func forwardDownstream(client *Client, cmd interface{}, logger *zap.Logger) (interface{}, error) {
	remote := client.streams[0]
	res, err := remote.sendReceive([]interface{}{[]byte{'p', 'i', 'n', 'g'}, []byte{'1', '2', '3'}}, logger)
	if err != nil {
		logger.Error("Can't ping", zap.Error(err))
		return nil, err
	}
	if string(*res.(*[]byte)) != "123" {
		return nil, fmt.Errorf("Sequencing error got: '%v'", res)
	}
	return remote.sendReceive(cmd, logger)
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

	remoteConn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		s.logger.Error("Can't connect to real Redis", zap.Error(err))
		return
	}
	defer remoteConn.Close()
	remote := DownStream{
		reader: NewReader(remoteConn, s.logger),
		writer: NewWriter(remoteConn)}
	client.streams = append(make([]DownStream, 0), remote)

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
