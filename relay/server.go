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


func handleConnection(conn net.Conn, s* Server) {
	defer conn.Close()

	reader := NewReader(conn)

	for {
		cmd, err := reader.ScanCommand()
		if err != nil {
			// TODO(holgerf): Add a bug message here...
			return
		}
		fmt.Printf("%v...\n", cmd)
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
