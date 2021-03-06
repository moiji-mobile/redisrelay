package relay

import (
	"fmt"
	pb "github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"net"
	"strconv"
	"time"
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
	options *ServerOptions
	logger  *zap.Logger
	reader  *Reader
	writer  *Writer
	remotes []remote
}

// A result coming from downstream connection
type ForwardResult struct {
	result interface{}
	err    error
}

func (forward *ForwardResult) SetResultForTesting(result interface{}) {
	forward.result = result
}

func (forward ForwardResult) GetResultForTesting() (result interface{}) {
	return forward.result
}

func NewClient(options *ServerOptions, logger *zap.Logger) *Client {
	return &Client{options: options, logger: logger}
}

func NewServer(opt *ServerOptions) (*Server, error) {
	if len(opt.GetRemoteAddresses()) == 0 {
		return nil, fmt.Errorf("Need to have at least one downstream: %v", len(opt.GetRemoteAddresses()))
	}

	TimeOut, err := pb.Duration(opt.GetRequestTimeout())
	if err != nil {
		return nil, err
	}
	opt.TimeOut = TimeOut

	ln, err := net.Listen("tcp", opt.GetBindAddress())
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
	// TODO(zecke): Add proper pool handling. This will require to re-try in case
	// of a connection failure.
	if stream != nil && stream.conn != nil {
		stream.conn.Close()
	}
}

func (remote *remote) forwardCommand(c chan<- ForwardResult, cmd interface{}, logger *zap.Logger) {
	stream, err := remote.getDownStream()
	if err != nil {
		c <- ForwardResult{result: nil, err: err}
	} else {
		res, err := stream.sendReceive(cmd, logger)
		c <- ForwardResult{result: res, err: err}
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

func FindVersion(result interface{}, version string) int64 {
	// Is this an array?
	arr, valid := result.([]interface{})
	if !valid {
		return 0
	}

	// An array of alternating key/value. Find the key and then
	// cast/extract the value.
	for i := 0; i < len(arr); i += 2 {
		name := GetString(arr[i])
		if name != nil && *name == version && i+1 < len(arr) {
			if v, ok := arr[i+1].(int64); ok {
				return v
			}
			if v, ok := arr[i+1].([]byte); ok {
				v, err := strconv.ParseInt(string(v), 10, 64)
				if err == nil {
					return v
				}
			}
			if v, ok := arr[i+1].(*[]byte); ok {
				v, err := strconv.ParseInt(string(*v), 10, 64)
				if err == nil {
					return v
				}
			}
			return 0
		}
	}
	return 0
}

func (client *Client) SelectBestResult(results []ForwardResult, resp_has_ver bool) (interface{}, error) {
	if !resp_has_ver {
		return results[0].result, results[0].err
	}

	var selected ForwardResult
	var max *int64

	for _, result := range results {
		dataVersion := FindVersion(result.result, client.options.GetVersionFieldName())
		if max == nil || dataVersion > *max {
			selected = result
			max = &dataVersion
		}
	}
	return selected.result, selected.err
}

func (client *Client) SelectResult(results []ForwardResult, errors []ForwardResult, resp_has_ver bool) (interface{}, error) {
	// Pick any success and then any error.
	if uint32(len(results)) >= client.options.GetMinSuccess() {
		return client.SelectBestResult(results, resp_has_ver)
	}
	if len(errors) > 0 {
		return errors[0].result, errors[0].err
	}
	return nil, fmt.Errorf("Not enough success nor failure. Reporting as failed")
}

func DetermineResponseWillHaveVersion(cmd interface{}) bool {
	// Is this an array?
	arr, valid := cmd.([]interface{})
	if !valid {
		return false
	}

	// Does it hold enough elements?
	if len(arr) != 2 {
		return false
	}

	// Is the first one HGETALL?
	name, valid := arr[0].(*[]byte)
	if !valid {
		return false
	}

	if string(*name) != "HGETALL" {
		return false
	}

	return true
}

func forwardDownstream(client *Client, cmd interface{}, logger *zap.Logger) (interface{}, error) {
	// Is this a request that will yield a response with a version inside?
	resp_has_ver := DetermineResponseWillHaveVersion(cmd)

	c := make(chan ForwardResult, len(client.remotes))
	failures := make([]ForwardResult, 0, len(client.remotes))
	results := make([]ForwardResult, 0, len(client.remotes))

	for _, remote := range client.remotes {
		var aremote = remote
		aremote.forwardCommand(c, cmd, logger)
	}

	// Start the timer after we have queued all requests.

	timeOut := time.NewTimer(client.options.TimeOut)
	for _, _ = range client.remotes {
		select {
		case <-timeOut.C:
			client.logger.Error("Time out getting a request")
			return client.SelectResult(results, failures, resp_has_ver)
		case f := <-c:
			if f.err != nil {
				failures = append(failures, f)
			} else {
				results = append(results, f)
			}
		}
	}
	timeOut.Stop()
	return client.SelectResult(results, failures, resp_has_ver)
}

func (client *Client) forwardCommands() {
	for {
		// Get the command
		cmd, err := client.reader.ParseCommand()
		if err != nil {
			client.logger.Error("Can't parse command", zap.Error(err))
			return
		}

		res, err := forwardDownstream(client, cmd, client.logger)
		if err != nil {
			client.logger.Error("Can't forward command", zap.Error(err))
			return
		}

		// Write the response
		err = client.writer.Write(res)
		if err != nil {
			client.logger.Error("Can't forward response", zap.Error(err))
			return
		}
		client.writer.Flush()
	}
}

func handleConnection(conn net.Conn, s *Server) {
	defer conn.Close()

	client := Client{
		options: s.options,
		logger:  s.logger,
		reader:  NewReader(conn, s.logger),
		writer:  NewWriter(conn)}

	client.remotes = make([]remote, 0)
	for _, addr := range s.options.GetRemoteAddresses() {
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
