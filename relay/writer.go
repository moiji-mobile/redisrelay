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
	err := w.BufioWriter.WriteByte('*')
	if err != nil {
		return err
	}

	err = w.writeInt(int64(len(vals)))
	if err != nil {
		return err
	}

	for _, val := range vals {
		err := w.Write(val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeErrorString(val error) error {
	err := w.BufioWriter.WriteByte('-')
	if err != nil {
		return err
	}

	_, err = w.BufioWriter.WriteString(val.Error())
	if err != nil {
		return err
	}
	return w.crlf()
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
	err := w.BufioWriter.WriteByte('$')
	if err != nil {
		return err
	}

	if val == nil {
		return w.writeInt(int64(-1))
	}

	err = w.writeInt(int64(len(*val)))
	if err != nil {
		return err
	}

	_, err = w.BufioWriter.Write(*val)
	if err != nil {
		return err
	}
	return w.crlf()
}

func (w *Writer) writeSimpleString(val SimpleString) error {
	err := w.BufioWriter.WriteByte('+')
	if err != nil {
		return err
	}
	_, err = w.BufioWriter.WriteString(val.String)
	if err != nil {
		return err
	}
	return w.crlf()
}

func (w *Writer) crlf() (err error) {
	_, err = w.BufioWriter.Write([]byte{'\r', '\n'})
	return
}

func (w *Writer) Write(val interface{}) (err error) {
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

	return
}

func (w *Writer) Flush() error {
	return w.BufioWriter.Flush()
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{BufioWriter: bufio.NewWriter(conn)}
}
