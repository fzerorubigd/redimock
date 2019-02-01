package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadArray(t *testing.T) {
	type cas struct {
		payload string
		err     error
		res     []string
	}
	for _, c := range []cas{
		{
			payload: "*1\r\n$4\r\nPING\r\n",
			res:     []string{"PING"},
		},
		{
			payload: "*2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n",
			res:     []string{"LLEN", "mylist"},
		},
		{
			payload: "*2\r\n$4\r\nLLEN\r\n$6\r\nmyl",
			err:     io.EOF,
		},
		{
			payload: "PING",
			err:     io.EOF,
		},
		{
			payload: "*0\r\n",
		},
		{
			payload: "*-1\r\n", // not sure this is legal in a request
		},
		{
			payload: "\r\n",
			err:     ErrProtocol,
		},
		{
			payload: "*NO\r\n",
			err:     ErrProtocol,
		},
		{
			payload: "&10\r\n",
			err:     ErrProtocol,
		},
	} {
		res, err := readArray(bytes.NewBufferString(c.payload))
		assert.Equal(t, err, c.err)
		assert.Equal(t, res, c.res)
	}
}

func TestReadString(t *testing.T) {
	type cas struct {
		payload string
		err     error
		res     string
	}
	bigPayload := strings.Repeat("X", 1<<24)
	for _, c := range []cas{
		{
			payload: "+hello world\r\n",
			res:     "hello world",
		},
		{
			payload: "-some error\r\n",
			res:     "some error",
		},
		{
			payload: ":42\r\n",
			res:     "42",
		},
		{
			payload: ":\r\n",
			res:     "",
		},
		{
			payload: "$4\r\nabcd\r\n",
			res:     "abcd",
		},
		{
			payload: fmt.Sprintf("$%d\r\n%s\r\n", len(bigPayload), bigPayload),
			res:     bigPayload,
		},

		{
			payload: "",
			err:     io.EOF,
		},
		{
			payload: ":42",
			err:     io.EOF,
		},
		{
			payload: "XXX",
			err:     io.EOF,
		},
		{
			payload: "XXXXXX",
			err:     io.EOF,
		},
		{
			payload: "\r\n",
			err:     ErrProtocol,
		},
		{
			payload: "XXXX\r\n",
			err:     ErrProtocol,
		},
		{
			payload: "$HI\r\n",
			err:     ErrProtocol,
		},
		{
			payload: "$-1\r\n",
			res:     "",
		},
		{
			payload: "$100\r\nNO\r\n",
			err:     io.EOF,
		},
	} {
		res, err := readString(bufio.NewReader(bytes.NewBufferString(c.payload)))
		assert.Equal(t, err, c.err)
		assert.Equal(t, res, c.res)
	}
}

type failWriter int

func (failWriter) Write(p []byte) (n int, err error) {
	return 0, io.EOF
}

func TestProtoWrite(t *testing.T) {
	w := &bytes.Buffer{}
	assert.Error(t, write(w, map[string]string{}))
	assert.Error(t, write(failWriter(0), []int{1, 2, 3}))
}
