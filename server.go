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
			panic(err)
		}

		if cmd.terminate {
			return
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
func (s *Server) Expect(command string, args ...string) *Command {
	c := &Command{
		command: strings.ToUpper(command),
	}
	c.argCompare = func(s ...string) bool {
		if len(s) != len(args) {
			return false
		}

		for i := range s {
			if s[i] != args[i] {
				return false
			}
		}
		return true
	}
	s.expectList = append(s.expectList, c)

	return c
}

// ExpectWithAnyArg is like expect but accept any argument
func (s *Server) ExpectWithAnyArg(command string) *Command {
	c := s.Expect(command)
	c.argCompare = func(...string) bool {
		return true
	}
	return c
}

// ExpectWithFn is the custom except with function
func (s *Server) ExpectWithFn(command string, fn func(...string) bool) *Command {
	c := s.Expect(command)
	c.argCompare = fn
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
	if len(input) < 1 {
		return false
	}

	if strings.ToUpper(input[0]) != c.command {
		return false
	}

	if c.argCompare != nil {
		return c.argCompare(input[1:]...)
	}

	return false
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
