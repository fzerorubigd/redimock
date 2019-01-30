package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Command is a single
type Command struct {
	command   string
	args      []string
	responses []interface{}
	count     int
	terminate bool
	delay     time.Duration

	lock   sync.Mutex
	called int
}

// Server is the mock server used for handling the connections
type Server struct {
	listener net.Listener
	lock     sync.RWMutex

	expectList         []*Command
	unexpectedCommands [][]string
}

// NewServer makes a server listening on addr. Close with .Close().
func NewServer(ctx context.Context, addr string) (*Server, error) {
	s := Server{}
	lc := net.ListenConfig{}
	l, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	s.listener = l
	go s.serve()

	return &s, nil
}

func (s *Server) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.ServeConn(conn)
	}
}

// ServeConn handles a net.Conn. Nice with net.Pipe()
func (s *Server) ServeConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	r := bufio.NewReader(conn)
	for {
		args, err := readArray(r)
		if err != nil {
			panic(err)
		}
		var cmd *Command
		for i := range s.expectList {
			if s.expectList[i].compare(args) {
				cmd = s.expectList[i]
				cmd.increase()
			}
		}
		if cmd == nil {
			// Return error *and continue?*
			if err := writeError(conn, ""); err != nil {
				panic(err)
			}
			s.unexpectedCommands = append(s.unexpectedCommands, args)
			continue
		}

		if cmd.delay > 0 {
			time.Sleep(cmd.delay)
		}
		if err := write(conn, cmd.responses...); err != nil {
			panic(err)
		}

		if cmd.terminate {
			return
		}
	}
}

// Addr has the net.Addr struct
func (s *Server) Addr() *net.TCPAddr {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.listener == nil {
		return nil
	}
	return s.listener.Addr().(*net.TCPAddr)
}

// Close closes all connections
func (s *Server) Close() error {
	s.lock.Lock()
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return err
		}
	}
	s.listener = nil
	s.lock.Unlock()
	return nil
}

// ExpectationsWereMet return nil if the all expects match or error if not
func (s *Server) ExpectationsWereMet() error {
	var all []error
	for i := range s.expectList {
		if err := s.expectList[i].error(); err != nil {
			all = append(all, err)
		}
	}

	for i := range s.unexpectedCommands {
		all = append(all, fmt.Errorf(
			"command %s is called but not expected",
			strings.Join(s.unexpectedCommands[i], " ")),
		)
	}

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
func (s *Server) Expect(command string, args ...string) *Command {
	c := &Command{
		command: strings.ToUpper(command),
		args:    args,
	}
	s.expectList = append(s.expectList, c)

	return c
}

// WillReturn set the return value for this command
func (c *Command) WillReturn(ret ...interface{}) *Command {
	c.responses = ret

	return c
}

func (c *Command) WithDelay(d time.Duration) *Command {
	c.delay = d
	return c
}

// Once means it should be called once
func (c *Command) Once() *Command {
	c.count = 1
	return c
}

func (c *Command) Any() *Command {
	c.count = -1
	return c
}

func (c *Command) Times(n int) *Command {
	c.count = n
	return c
}

func (c *Command) CloseConnection() *Command {
	c.terminate = true
	return c
}

func (c *Command) compare(input []string) bool {
	if len(input) != len(c.args)+1 {
		return false
	}
	if strings.ToUpper(input[0]) != c.command {
		return false
	}

	for i := range c.args {
		if c.args[i] != input[i+1] {
			return false
		}
	}

	return true
}

func (c *Command) error() error {
	if c.count < 0 || c.count == c.called {
		return nil
	}

	return fmt.Errorf(
		`command "%s %s" expected %d time called %d times`,
		c.command,
		strings.Join(c.args, " "),
		c.count,
		c.called,
	)
}

func (c *Command) increase() {
	c.lock.Lock()
	c.called++
	c.lock.Unlock()
}
