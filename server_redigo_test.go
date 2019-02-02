package redimock

import (
	"context"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"
)

func TestRedigoQuit(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	s.ExpectQuit().Once()

	red, err := redis.Dial("tcp", s.Addr().String())
	require.NoError(t, err)

	st, err := redis.String(red.Do("QUIT"))
	require.NoError(t, err)
	require.Equal(t, "OK", st)

	require.NoError(t, s.ExpectationsWereMet())
}

func TestRedigoStringKey(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	s.ExpectSet("test_key", "Hello", true).Once()
	s.ExpectSet("test_key", "Hello2", false, "ex", "1", "nx").Once()
	s.ExpectSet("test_key", "Hello2", false).Once()

	s.ExpectGet("test_key", true, "Hello").Once()
	s.ExpectGet("test_key_2", false, "").Once()

	red, err := redis.Dial("tcp", s.Addr().String())
	require.NoError(t, err)

	x, e := redis.String(red.Do("SET", "test_key", "Hello"))
	require.NoError(t, e)
	require.Equal(t, "OK", x)

	_, e = redis.String(red.Do("SET", "test_key", "Hello2", "ex", "1", "nx"))
	require.Equal(t, redis.ErrNil, e)

	x, e = redis.String(red.Do("SET", "test_key", "Hello2"))
	require.NoError(t, e)
	require.Equal(t, "OK", x)

	x, e = redis.String(red.Do("get", "test_key"))
	require.NoError(t, e)
	require.Equal(t, "Hello", x)

	_, e = redis.String(red.Do("get", "test_key_2"))
	require.Error(t, e)

	require.NoError(t, s.ExpectationsWereMet())
}

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

	require.NoError(t, s.ExpectationsWereMet())
}

func TestRedigoListHGetAll(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	v := map[string]string{"v1": "v2", "v3": "v4"}
	s.ExpectHGetAll("mykey", v).Once()

	red, err := redis.Dial("tcp", s.Addr().String())
	require.NoError(t, err)

	ret, err := redis.StringMap(red.Do("hgetall", "mykey"))
	require.NoError(t, err)
	require.Equal(t, v, ret)

	require.NoError(t, s.ExpectationsWereMet())
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

	require.Error(t, s.ExpectationsWereMet())
}

func TestRedigoLPush(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	s.ExpectLPush(2, "LIST", "KEY1", "KEY2").Once()
	s.ExpectLPush(5, "LIST").Once()
	s.ExpectRPush(6, "LIST").Once()

	red, err := redis.Dial("tcp", s.Addr().String())
	require.NoError(t, err)

	ret, err := redis.Int(red.Do("lpush", "LIST", "KEY1", "KEY2"))
	require.NoError(t, err)
	require.Equal(t, 2, ret)

	ret, err = redis.Int(red.Do("lpush", "LIST", "HI", "HOY", "BOY"))
	require.NoError(t, err)
	require.Equal(t, 5, ret)

	ret, err = redis.Int(red.Do("rpush", "LIST", "R"))
	require.NoError(t, err)
	require.Equal(t, 6, ret)

	require.NoError(t, s.ExpectationsWereMet())

	_, err = redis.Int(red.Do("lpush"))
	require.Error(t, err)
	require.Error(t, s.ExpectationsWereMet())
}
