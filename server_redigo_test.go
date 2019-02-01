package server

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRedigoList(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	v := map[string]string{"v1": "v2", "v3": "v4"}
	s.ExpectHGetAll("mykey", v)

	red, err := redis.Dial("tcp", s.Addr().String())
	require.NoError(t, err)

	ret, err := redis.StringMap(red.Do("hgetall", "mykey"))
	require.NoError(t, err)
	require.Equal(t, v, ret)
}
