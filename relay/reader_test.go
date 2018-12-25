package relay_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/moiji-mobile/redisrelay/relay"
	"go.uber.org/zap"
)

func newReader(str string) *relay.Reader {
	logger, _ := zap.NewDevelopment()
	return relay.NewReader(strings.NewReader(str), logger)
}

func parseCommand(inp string, t *testing.T) []interface{} {
	r := newReader(inp)
	cmd, err := r.ParseCommand()
	if err != nil {
		t.Errorf("Failed with err=%v on inp=%v\n", err, inp)
	}
	return cmd
}

func parseData(inp string, t *testing.T) interface{} {
	r := newReader(inp)
	cmd, err := r.ParseData()
	if err != nil {
		t.Errorf("Failed with err=%v on inp=%v\n", err, inp)
	}
	return cmd
}

func testRoundTrip(cmd interface{}, inp string, t *testing.T) {
	b := bytes.NewBuffer(make([]byte, 0, len(inp)))
	w := relay.NewWriter(b)
	err := w.Write(cmd)
	w.BufioWriter.Flush()
	if err != nil {
		t.Errorf("Unpexted error: %v\n", err)
	}
	res := string(b.Bytes())
	if inp != res {
		t.Errorf("Wanted '%v' but got '%v'", inp, res)
	}
}

func TestParseCommand_ArrayWithBulkString(t *testing.T) {
	inp := "*2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n"
	res := parseCommand(inp, t)
	if len(res) != 2 {
		t.Errorf("Expected a length of two but got %v\n", len(res))
	}
	if string(*res[0].(*[]byte)) != "LLEN" {
		t.Errorf("Expected LLEN but got %v\n", res[0].(*[]byte))
	}
	if string(*res[1].(*[]byte)) != "mylist" {
		t.Errorf("Expected nmylist but got %v\n", res[1].(*[]byte))
	}
	testRoundTrip(res, inp, t)
}

func TestParseCommand_ArrayEmpty(t *testing.T) {
	inp := "*0\r\n"
	res := parseCommand(inp, t)
	testRoundTrip(res, inp, t)
}

func TestParseCommand_ArrayMixed(t *testing.T) {
	inp := "*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nfoobar\r\n"
	res := parseCommand(inp, t)
	testRoundTrip(res, inp, t)
}

func TestParseData_BulkStringNull(t *testing.T) {
	inp := "$-1\r\n"
	res := parseData(inp, t)
	if res.(*[]byte) != nil {
		t.Errorf("Wanted null string but got: %v %v", res, res == nil)
	}
	testRoundTrip(res, inp, t)
}

func TestParseData_BulkStringEmpty(t *testing.T) {
	inp := "$0\r\n\r\n"
	res := parseData(inp, t)
	if string(*res.(*[]byte)) != "" {
		t.Errorf("Wanted empty string but got: %v %v", res, res == nil)
	}
	testRoundTrip(res, inp, t)
}

func TestParseData_SimpleString(t *testing.T) {
	inp := "+OK\r\n"
	res := parseData(inp, t)
	if res.(relay.SimpleString).String != "OK" {
		t.Errorf("Wanted OK string but got: %v", string(res.([]byte)))
	}
	testRoundTrip(res, inp, t)
}

func TestParseData_Error(t *testing.T) {
	inp := "-Error message\r\n"
	res := parseData(inp, t)
	if res.(error).Error() != "Error message" {
		t.Errorf("Wanted 'Error message' string but got: %v", string(res.([]byte)))
	}
	testRoundTrip(res, inp, t)
}

func TestGetString_String(t *testing.T) {
	inp := []byte{'s', 't', 'r', 'i', 'n', 'g'}
	inp_ptr := &inp

	// Positive tests
	res := relay.GetString(inp)
	if *res != string(inp) {
		t.Errorf("Couldn't extract string. Got: %#v", res)
	}
	res = relay.GetString(inp_ptr)
	if *res != string(inp) {
		t.Errorf("Couldn't extract string. Got: %#v", res)
	}

	res = relay.GetString(nil)
	if res != nil {
		t.Errorf("Expected a nil ptr")
	}
	res = relay.GetString(3)
	if res != nil {
		t.Errorf("Expected a nil ptr")
	}
}
