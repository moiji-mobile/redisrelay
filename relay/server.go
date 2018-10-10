package relay

import (
	"net"
	"go.uber.org/zap"
)

type Server struct {
	listener net.Listener
	logger *zap.Logger
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

func handleConnection(conn net.Conn, s *Server) {
	defer conn.Close()

	reader := NewReader(conn, s.logger)
	writer := NewWriter(conn)

	remote, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		s.logger.Error("Can't connect to real Redis", zap.Error(err))
		return
	}
	defer remote.Close()
	remoteReader := NewReader(remote, s.logger)
	remoteWriter := NewWriter(remote)

	for {
		cmd, err := reader.ParseCommand()
		if err != nil {
			s.logger.Error("Can't parse command", zap.Error(err))
			return
		}

		err = remoteWriter.Write(cmd)
		if err != nil {
			s.logger.Error("Can't forward command", zap.Error(err))
			return
		}
		remoteWriter.Flush()

		res, err := remoteReader.ParseData()
		if err != nil {
			s.logger.Error("Can't response", zap.Error(err))
			return
		}

		err = writer.Write(res)
		if err != nil {
			s.logger.Error("Can't forward response", zap.Error(err))
			return
		}
		writer.Flush()
	}
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
