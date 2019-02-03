package redimock

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// Command is a single
type Command struct {
	command    string
	argCompare func(...string) bool
	responses  []interface{}
	count      int
	terminate  bool
	delay      time.Duration

	lock   sync.Mutex
	called int
}

// Result is the function that can be used for advanced result value
type Result = func(...string) []interface{}

// Server is the mock server used for handling the connections
type Server struct {
	listener net.Listener

	expectList         []*Command
	lock               sync.RWMutex
	unexpectedCommands [][]string
}

// NewServer makes a server listening on addr. Close with .Close().
func NewServer(ctx context.Context, addr string) (*Server, error) {
	s := Server{}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s.listener = l
	go s.serve()
	go func() {
		<-ctx.Done()
		_ = l.Close()
	}()
	return &s, nil
}

func (s *Server) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go func() {
			// TODO : add this err to log?
			_ = s.serveConn(conn)
		}()
	}
}

// ServeConn handles a connection
func (s *Server) serveConn(conn io.ReadWriteCloser) error {
	defer func() {
		_ = conn.Close()
	}()
	for {
		args, err := readArray(conn)
		if err != nil {
			// Close the connection and return, error in client should not break the server
			return err
		}

		var cmd *Command
		for i := range s.expectList {
			if s.expectList[i].compare(args) {
				cmd = s.expectList[i]
				cmd.increase()
				break
			}
		}

		if cmd == nil {
			// Return error *and continue?*
			if err := write(conn, Error("command not expected")); err != nil {
				return err // this means the write was not successful , close the connection
			}
			s.lock.Lock()
			s.unexpectedCommands = append(s.unexpectedCommands, args)
			s.lock.Unlock()
			continue
		}

		if cmd.delay > 0 {
			time.Sleep(cmd.delay)
		}

		rsp := cmd.responses
		if len(rsp) == 1 {
			fn, ok := rsp[0].(Result)
			if ok {
				rsp = fn(args[1:]...)
			}
		}
		if err := write(conn, rsp...); err != nil {
			// write failed, return and close the connection
			return err
		}

		if cmd.terminate {
			return nil
		}
	}
}

// Addr has the net.Addr struct
func (s *Server) Addr() *net.TCPAddr {
	return s.listener.Addr().(*net.TCPAddr)
}

// ExpectationsWereMet return nil if the all expects match or error if not
func (s *Server) ExpectationsWereMet() error {
	var all []error
	for i := range s.expectList {
		if err := s.expectList[i].error(); err != nil {
			all = append(all, err)
		}
	}

	s.lock.RLock()
	for i := range s.unexpectedCommands {
		all = append(all, fmt.Errorf(
			"command %s is called but not expected",
			strings.Join(s.unexpectedCommands[i], " ")),
		)
	}
	s.lock.RUnlock()

	var str string
	if len(all) > 0 {
		for i := range all {
			str += all[i].Error() + "\n"
		}
		return fmt.Errorf(str)
	}

	return nil
}

// Expect return a command
func (s *Server) Expect(command string) *Command {
	c := &Command{
		command: strings.ToUpper(command),
	}

	s.expectList = append(s.expectList, c)
	return c
}

// WithArgs add array as arguments
func (c *Command) WithArgs(args ...string) *Command {
	return c.WithFnArgs(func(s ...string) bool {
		if len(s) != len(args) {
			return false
		}

		return equalArgs(s, args)
	})
}

// WithAnyArgs if any argument is ok
func (c *Command) WithAnyArgs() *Command {
	return c.WithFnArgs(func(...string) bool {
		return true
	})
}

// WithFnArgs is advanced function compare for arguments
func (c *Command) WithFnArgs(f func(...string) bool) *Command {
	// TODO : may be panic() if the function already set
	c.argCompare = f
	return c
}

// WillReturn set the return value for this command
func (c *Command) WillReturn(ret ...interface{}) *Command {
	c.responses = ret

	return c
}

// WillReturnFn is a helper function for advance return value
func (c *Command) WillReturnFn(r Result) *Command {
	return c.WillReturn(r)
}

// WithDelay return command with delay
func (c *Command) WithDelay(d time.Duration) *Command {
	c.delay = d
	return c
}

// Once means it should be called once
func (c *Command) Once() *Command {
	c.count = 1
	return c
}

// Any means this can be called 0 to n time
func (c *Command) Any() *Command {
	c.count = -1
	return c
}

// Times this should be called n times
func (c *Command) Times(n int) *Command {
	c.count = n
	return c
}

// CloseConnection should close connection after this command
func (c *Command) CloseConnection() *Command {
	c.terminate = true
	return c
}

func (c *Command) compare(input []string) bool {
	if len(input) < 1 {
		return false
	}

	if strings.ToUpper(input[0]) != c.command {
		return false
	}

	if c.argCompare != nil {
		return c.argCompare(input[1:]...)
	}

	return len(input) == 1
}

func (c *Command) error() error {
	if c.count < 0 || c.count == c.called {
		return nil
	}

	return fmt.Errorf(
		`command "%s" expected %d time called %d times`,
		c.command,
		c.count,
		c.called,
	)
}

func (c *Command) increase() {
	c.lock.Lock()
	c.called++
	c.lock.Unlock()
}
