package relay_test

import (
	"github.com/moiji-mobile/redisrelay/relay"
	"strings"
	"testing"
)

func newReader(str string) *relay.Reader {
	return relay.NewReader(strings.NewReader(str))
}


func checkResult(inp string, err error, cmd []byte, t *testing.T) {
	if err != nil {
		t.Errorf("Parsing resulted in error: %v", err)
	}
	if string(cmd) != inp {
		t.Errorf("Strings don't match: '%v' vs. '%v'", string(cmd), inp)
	}
}


func roundTripTestScanCommand(inp string, t *testing.T) {
	r := newReader(inp)
	cmd, err := r.ScanCommand()
	checkResult(inp, err, cmd, t)
}

func roundTripTestScanData(inp string, t *testing.T) {
	r := newReader(inp)
	cmd, err := r.ScanData()
	checkResult(inp, err, cmd, t)
}

func TestScanCommand_ArrayWithBulkString(t *testing.T) {
	inp := "*2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n"
	roundTripTestScanCommand(inp, t)
}

func TestScanCommand_ArrayEmpty(t *testing.T) {
	inp := "*0\r\n"
	roundTripTestScanCommand(inp, t)
}

func TestScanCommand_ArrayMixed(t *testing.T) {
	inp := "*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nfoobar\r\n"
	roundTripTestScanCommand(inp, t)
}

func TestScanData_BulkStringNull(t *testing.T) {
	inp := "$-1\r\n"
	roundTripTestScanData(inp, t)
}

func TestScanData_BulkStringEmpty(t *testing.T) {
	inp := "$0\r\n\r\n"
	roundTripTestScanData(inp, t)
}

func TestScanData_SimpleString(t *testing.T) {
	inp := "+OK\r\n"
	roundTripTestScanData(inp, t)
}

func TestScanData_Error(t *testing.T) {
	inp := "-Error message\r\n"
	roundTripTestScanData(inp, t)
}
