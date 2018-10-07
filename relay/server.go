package relay

import (
	"fmt"
	"net"
)

type Server struct {
	listener net.Listener
}

func NewServer(opt *ServerOptions) (*Server, error) {
	opt.init()

	ln, err := net.Listen("tcp", opt.Address)
	if err != nil {
		return nil, err
	}

	s := Server{listener: ln}
	return &s, nil
}

func handleConnection(conn net.Conn, s *Server) {
	defer conn.Close()

	reader := NewReader(conn)
	writer := NewWriter(conn)

	remote, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		return
	}
	defer remote.Close()
	remoteReader := NewReader(remote)
	remoteWriter := NewWriter(remote)

	for {
		cmd, err := reader.ParseCommand()
		fmt.Println(cmd)
		if err != nil {
			// TODO(holgerf): Add a bug message here...
			return
		}

		err = remoteWriter.Write(cmd)
		if err != nil {
			// TODO(holgerf): ...
			return
		}
		remoteWriter.Flush()

		res, err := remoteReader.ParseData()
		if err != nil {
			// TODO(holgerf): ...
			return
		}

		err = writer.Write(res)
		if err != nil {
			// TODO(holgerf): ...
			return
		}
		writer.Flush()
	}
}

func (s *Server) Run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// TODO(zecke): Log things...
		}
		go handleConnection(conn, s)
	}
}
