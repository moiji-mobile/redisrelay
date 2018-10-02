package relay

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Reader struct {
	BufioReader *bufio.Reader
}

func (r *Reader) ScanLine() ([]byte, error) {
	line, _, err := r.BufioReader.ReadLine()

	if err == nil {
		return line, err
	}

	// TODO(holgerf): Improve error handling.
	fmt.Printf("Got err '%v' %v\n", string(line), err)
	return nil, err
}

// Scan an array.
func (r *Reader) ScanArray() ([]byte, error) {
	line, err := r.ScanLine()

	if err != nil {
		// TODO(holgerf):
		return nil, err
	}

	// Is this an array?
	if line[0] != '*' {
		return nil, fmt.Errorf("Not an array.")
	}

	return r.ScanArrayElements(line)
}

func (r *Reader) ScanArrayElements(line []byte) ([]byte, error) {
	// How many entries to parse?
	l, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		return nil, err
	}

	line = append(line, "\r\n"...)
	// Collect all of the lines.
	for i := int64(0); i < l; i++ {
		extra, err := r.ScanData()
		if err != nil {
			return nil, err
		}
		line = append(line, extra...)
	}
	return line, nil
}

func (r *Reader) ScanSimpleString(line []byte) ([]byte, error) {
	return append(line, "\r\n"...), nil
}

func (r *Reader) ScanIntegers(line []byte) ([]byte, error) {
	_, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		return nil, err
	}
	return append(line, "\r\n"...), nil
}

func (r *Reader) ScanBulkString(line []byte) ([]byte, error) {
	l, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		return nil, err
	}

	// $-1\r\n indicates a NULL string.
	if l < 0 {
		line = append(line, "\r\n"...)
		return line, nil
	}

	buf := make([]byte, l+2)
	_, err = io.ReadFull(r.BufioReader, buf)
	if err != nil {
		return nil, err
	}

	line = append(line, "\r\n"...)
	line = append(line, buf...)
	return line, nil
}

func (r *Reader) ScanData() ([]byte, error) {
	line, err := r.ScanLine()

	if err != nil {
		return nil, err
	}

	switch line[0] {
	case '+':
		return r.ScanSimpleString(line)
	case '-':
		return r.ScanSimpleString(line)
	case ':':
		return r.ScanIntegers(line)
	case '$':
		return r.ScanBulkString(line)
	case '*':
		return r.ScanArrayElements(line)
	default:
		return nil, fmt.Errorf("Can't parse data: %s", line)
	}
}

// Parse a single a Command. It must be an array.
func (r *Reader) ScanCommand() ([]byte, error) {
	return r.ScanArray()
}

func NewReader(conn io.Reader) *Reader {
	r := Reader{BufioReader: bufio.NewReader(conn)}
	return &r
}
