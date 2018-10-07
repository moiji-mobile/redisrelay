package relay

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Writer struct {
	BufioWriter *bufio.Writer
}

func (w *Writer) writeArray(vals []interface{}) error {
	w.BufioWriter.WriteByte('*')
	w.writeInt(int64(len(vals)))

	for _, val := range vals {
		err := w.Write(val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeErrorString(val error) error {
	w.BufioWriter.WriteByte('-')
	w.BufioWriter.WriteString(val.Error())
	return w.crlf()
	return nil
}

func (w *Writer) writeInt(val int64) error {
	s := strconv.FormatInt(val, 10)
	_, err := w.BufioWriter.WriteString(s)
	if err != nil {
		return err
	}
	return w.crlf()
}

func (w *Writer) writeBulkString(val *[]byte) error {
	w.BufioWriter.WriteByte('$')

	if val == nil {
		return w.writeInt(int64(-1))
	}

	w.writeInt(int64(len(*val)))
	_, err := w.BufioWriter.Write(*val)
	if err != nil {
		return err
	}
	return w.crlf()
}

func (w *Writer) writeSimpleString(val SimpleString) error {
	w.BufioWriter.WriteByte('+')
	w.BufioWriter.WriteString(val.String)
	return w.crlf()
}

func (w *Writer) crlf() (err error) {
	_, err = w.BufioWriter.Write([]byte{'\r', '\n'})
	return
}

func (w *Writer) Write(val interface{}) error {
	var err error

	switch val.(type) {
	case []interface{}:
		err = w.writeArray(val.([]interface{}))
		break
	case *[]byte:
		err = w.writeBulkString(val.(*[]byte))
		break
	case []byte:
		str := val.([]byte)
		err = w.writeBulkString(&str)
		break
	case SimpleString:
		err = w.writeSimpleString(val.(SimpleString))
		break
	case int64:
		w.BufioWriter.WriteByte(':')
		err = w.writeInt(val.(int64))
		break
	case error:
		err = w.writeErrorString(val.(error))
		break
	default:
		err = fmt.Errorf("Type '%T' not handled.", val)
	}

	return err
}

func (w *Writer) Flush() error {
	return w.BufioWriter.Flush()
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{BufioWriter: bufio.NewWriter(conn)}
}
