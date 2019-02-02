package redimock

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGoRedisServerExpectationNotMet(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, ":12345")
	require.NoError(t, err)

	s.ExpectGet("abcd", true, "test").Once()

	require.Error(t, s.ExpectationsWereMet())
}

func TestNewServerErr(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	_, err := NewServer(ctx, ":INVALID")
	require.Error(t, err)
}

type closer struct {
	failOnWrite bool
	*bytes.Buffer
}

func (c *closer) Write(p []byte) (n int, err error) {
	if c.failOnWrite {
		return 0, io.EOF
	}
	return c.Buffer.Write(p)
}

func (closer) Close() error {
	return nil
}

func TestServerServeConErr(t *testing.T) {
	s := &Server{}
	buf := &closer{Buffer: bytes.NewBufferString("INVALID")}
	require.Panics(t, func() { s.serveConn(buf) })

	buf = &closer{Buffer: bytes.NewBufferString("*1\r\n$4\r\nPING\r\n"), failOnWrite: true,}
	require.Panics(t, func() { s.serveConn(buf) })

	s.Expect("PING").WithDelay(time.Microsecond).
		WillReturnFn(func(s ...string) []interface{} {
			return []interface{}{"OK"}
		}).Any()

	buf = &closer{Buffer: bytes.NewBufferString("*1\r\n$4\r\nPING\r\n"), failOnWrite: true,}
	require.Panics(t, func() { s.serveConn(buf) })

}

func TestCommandCompare(t *testing.T) {
	cmd := &Command{
		command: "PING",
	}

	require.False(t, cmd.compare(nil))
	require.True(t, cmd.compare([]string{"ping"}))
	require.False(t, cmd.compare([]string{"ping", "xxx"}))
	cmd.WithAnyArgs()
	require.True(t, cmd.compare([]string{"ping", "xxx"}))

}
