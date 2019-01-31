package server

import (
	"fmt"
	"strings"
)

// TODO : I don't want to fall in `implement another redis` trap. so be careful :)

// ExpectGet return a redis GET command
func (s *Server) ExpectGet(key string, exists bool, result string) *Command {
	c := s.Expect("GET", key)
	if exists {
		return c.WillReturn(BulkString(result))
	}
	return c.WillReturn(nil)
}

// ExpectSet return a redis set command. success could be false only for NX or XX option,
// otherwise it dose not make sense
func (s *Server) ExpectSet(key string, value string, success bool, extra ...string) *Command {
	args := append([]string{key, value}, extra...)
	c := s.Expect("SET", args...)
	if success {
		return c.WillReturn("OK")
	}
	return c.WillReturn(func(args ...string) []interface{} {
		for _, i := range args {
			x := strings.ToUpper(i)
			if x == "NX" || x == "XX" {
				return []interface{}{nil}
			}
		}
		return []interface{}{"OK"}
	})
}

// ExpectPing is the ping command
func (s *Server) ExpectPing() *Command {
	c := s.Expect("PING").WillReturn(func(args ...string) []interface{} {
		fmt.Println(args)
		if len(args) == 0 {
			return []interface{}{"PONG"}
		} else if len(args) == 1 {
			return []interface{}{args[0]}
		}
		return []interface{}{Error("ERR wrong number of arguments for 'ping' command")}
	})
	return c
}
