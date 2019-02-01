package server

import (
	"bufio"
)

/*
	This file is based on `https://github.com/alicebob/miniredis/blob/master/server/proto.go`
*/

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// ErrProtocol is the general error for unexpected input
var ErrProtocol = errors.New("invalid request")

type (
	// BulkString is used to handle the bulk string, normal strings are treated as simple string
	BulkString string
	// Error is the redis error type
	Error string
)

// client always sends arrays with bulk strings
func readArray(r io.Reader) ([]string, error) {
	rd := bufio.NewReader(r)
	line, err := rd.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if len(line) < 3 {
		return nil, ErrProtocol
	}

	switch line[0] {
	default:
		return nil, ErrProtocol
	case '*':
		l, err := strconv.Atoi(line[1 : len(line)-2])
		if err != nil {
			return nil, err
		}
		// l can be -1
		var fields []string
		for ; l > 0; l-- {
			s, err := readString(rd)
			if err != nil {
				return nil, err
			}
			fields = append(fields, s)
		}
		return fields, nil
	}
}

func readString(rd *bufio.Reader) (string, error) {
	line, err := rd.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 3 {
		return "", ErrProtocol
	}

	switch line[0] {
	default:
		return "", ErrProtocol
	case '+', '-', ':':
		// +: simple string
		// -: errors
		// :: integer
		// Simple line based replies.
		return string(line[1 : len(line)-2]), nil
	case '$':
		// bulk strings are: `$5\r\nhello\r\n`
		length, err := strconv.Atoi(line[1 : len(line)-2])
		if err != nil {
			return "", err
		}
		if length < 0 {
			// -1 is a nil response
			return "", nil
		}
		var (
			buf = make([]byte, length+2)
			pos = 0
		)
		for pos < length+2 {
			n, err := rd.Read(buf[pos:])
			if err != nil {
				return "", err
			}
			pos += n
		}
		return string(buf[:length]), nil
	}
}

func writeF(w io.Writer, s string, args ...interface{}) error {
	str := fmt.Sprintf(s, args...)
	_, err := fmt.Fprintf(w, str)
	if err != nil {
		return err
	}
	return nil
}

// writeError try to write a redis error to output
func writeError(w io.Writer, e Error) error {
	return writeF(w, "-%s\r\n", toInline(string(e)))
}

// writeSimpleString writes a redis inline string
func writeSimpleString(w io.Writer, s string) error {
	return writeF(w, "+%s\r\n", toInline(s))
}

// writeBulkString writes a bulk string
func writeBulkString(w io.Writer, s BulkString) error {
	return writeF(w, "$%d\r\n%s\r\n", len(s), s)
}

// writeNull writes a redis string NULL
func writeNull(w io.Writer) error {
	return writeF(w, "$-1\r\n")
}

// writeLen starts an array with the given length
func writeLen(w io.Writer, n int) error {
	return writeF(w, "*%d\r\n", n)
}

// writeInt writes an integer
func writeInt(w io.Writer, i int) error {
	return writeF(w, ":%d\r\n", i)
}

func toInline(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
}

func tryWriteArray(w io.Writer, t interface{}) error {
	// Now nasty reflection
	v := reflect.ValueOf(t)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("invalid type: %T", t)
	}

	l := v.Len()
	if err := writeLen(w, l); err != nil {
		return err
	}

	args := make([]interface{}, l)
	for i := range args {
		args[i] = v.Index(i).Interface()
	}

	return write(w, args...)
}

func write(w io.Writer, args ...interface{}) error {
	for i := range args {
		// first the easy way, no reflection
		switch t := args[i].(type) {
		case Error:
			// TODO : make sure its a one-liner
			if err := writeError(w, t); err != nil {
				return err
			}
		case BulkString:
			if err := writeBulkString(w, t); err != nil {
				return err
			}
		case int:
			if err := writeInt(w, t); err != nil {
				return err
			}
		case string:
			if err := writeSimpleString(w, t); err != nil {
				return err
			}
		case nil:
			if err := writeNull(w); err != nil {
				return err
			}
		default:
			if err := tryWriteArray(w, t); err != nil {
				return err
			}
		}
	}

	return nil
}
