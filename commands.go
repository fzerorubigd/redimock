package server

// ExpectGet return a redis GET command
func (s *Server) ExpectGet(key string, exists bool, result string) *Command {
	c := s.Expect("GET", key)
	if exists {
		return c.WillReturn(BulkString(result))
	}
	return c.WillReturn(nil)
}
