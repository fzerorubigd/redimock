package server

import (
	"context"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"
)

func TestRedigoPing(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	s.ExpectPing().Times(3)

	red, err := redis.Dial("tcp", s.Addr().String())
	require.NoError(t, err)

	st, err := redis.String(red.Do("PING"))
	require.NoError(t, err)
	require.Equal(t, "PONG", st)

	st, err = redis.String(red.Do("PING", "XXX"))
	require.NoError(t, err)
	require.Equal(t, "XXX", st)

	_, err = redis.String(red.Do("PING", "X", "Y"))
	require.Error(t, err)
}

func TestRedigoListHGetAll(t *testing.T) {
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

func TestRedigoListHSet(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	s.ExpectHSet("mykey", "fld", "value", true)

	red, err := redis.Dial("tcp", s.Addr().String())
	require.NoError(t, err)

	ret, err := redis.Int(red.Do("hset", "mykey", "fld", "value"))
	require.NoError(t, err)
	require.Equal(t, 0, ret)
}
