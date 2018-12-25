package relay

import (
	"bufio"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"strconv"
)

type Reader struct {
	bufioReader *bufio.Reader
	logger      *zap.Logger
}

func (r *Reader) ScanLine() ([]byte, error) {
	line, _, err := r.bufioReader.ReadLine()

	if err == nil {
		return line, err
	}

	return nil, err
}

// Parse an array.
func (r *Reader) ParseArray() ([]interface{}, error) {
	line, err := r.ScanLine()

	if err != nil {
		r.logger.Error("Failed to ReadLine", zap.Error(err))
		return nil, err
	}

	// Is this an array?
	if line[0] != '*' {
		return nil, fmt.Errorf("Not an array.")
	}

	return r.ParseArrayElements(line)
}

func (r *Reader) ParseArrayElements(line []byte) ([]interface{}, error) {
	// How many entries to parse?
	l, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		r.logger.Error("Failed to ParseInt", zap.Error(err), zap.ByteString("line", line[1:]))
		return nil, err
	}

	res := make([]interface{}, 0, l)
	// Collect all of the lines.
	for i := int64(0); i < l; i++ {
		extra, err := r.ParseData()
		if err != nil {
			r.logger.Error("Failed to parse element", zap.Error(err))
			return nil, err
		}
		res = append(res, extra)
	}
	return res, nil
}

func (r *Reader) ParseSimpleString(line []byte) (SimpleString, error) {
	return SimpleString{string(line[1:])}, nil
}

func (r *Reader) ParseSimpleError(line []byte) (error, error) {
	return errors.New(string(line[1:])), nil
}

func (r *Reader) ParseIntegers(line []byte) (int64, error) {
	num, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		r.logger.Error("Failed to parse int", zap.Error(err), zap.ByteString("line", line[1:]))
		return 0, err
	}
	return num, nil
}

func (r *Reader) ParseBulkString(line []byte) (*[]byte, error) {
	l, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		r.logger.Error("Failed to parse int", zap.Error(err), zap.ByteString("line", line[1:]))
		return nil, err
	}

	if l < 0 {
		return nil, nil
	}

	buf := make([]byte, l)
	_, err = io.ReadFull(r.bufioReader, buf)
	if err != nil {
		r.logger.Error("Failed to read", zap.Error(err))
		return nil, err
	}
	// Pop the \r\n that should be lingering around.
	_, err = r.ScanLine()
	if err != nil {
		r.logger.Error("Failed to ScanLine", zap.Error(err))
		return nil, err
	}

	return &buf, nil
}

func (r *Reader) ParseData() (interface{}, error) {
	line, err := r.ScanLine()

	if err != nil {
		return nil, err
	}

	switch line[0] {
	case '+':
		return r.ParseSimpleString(line)
	case '-':
		return r.ParseSimpleError(line)
	case ':':
		return r.ParseIntegers(line)
	case '$':
		return r.ParseBulkString(line)
	case '*':
		return r.ParseArrayElements(line)
	default:
		return nil, fmt.Errorf("Can't parse data: %s", line)
	}
}

// Parse a single a Command. It must be an array.
func (r *Reader) ParseCommand() ([]interface{}, error) {
	return r.ParseArray()
}

func NewReader(conn io.Reader, logger *zap.Logger) *Reader {
	r := Reader{bufioReader: bufio.NewReader(conn)}
	r.logger = logger
	return &r
}

func GetString(data interface{}) *string {
	if v, ok := data.([]uint8); ok {
		name := string(v)
		return &name
	}
	if v, ok := data.(*[]uint8); ok {
		name := string(*v)
		return &name
	}
	return nil
}
