package redimock

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
)

func TestGoRedisStringKey(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	s.ExpectSet("test_key", "Hello", true).Once()
	s.ExpectSet("test_key", "Hello2", false, "ex", "1", "nx").Once()

	s.ExpectGet("test_key", true, "Hello").Once()
	s.ExpectGet("test_key_2", false, "").Once()

	cl := redis.NewClient(&redis.Options{
		Addr: s.Addr().String(),
	})

	x, e := cl.Set("test_key", "Hello", 0).Result()
	require.NoError(t, e)
	require.Equal(t, "OK", x)

	b, e := cl.SetNX("test_key", "Hello2", time.Second).Result()
	require.NoError(t, e)
	require.False(t, b)

	str := cl.Get("test_key")
	x, e = str.Result()
	require.NoError(t, e)
	require.Equal(t, "Hello", x)

	str = cl.Get("test_key_2")
	_, e = str.Result()
	require.Error(t, e)

	require.NoError(t, s.ExpectationsWereMet())
}

func TestGoRedisExtraCall(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	cl := redis.NewClient(&redis.Options{
		Addr: s.Addr().String(),
	})

	cl.Ping()

	require.Error(t, s.ExpectationsWereMet())
}

func TestGoRedisPing(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	s.ExpectPing().Times(1)

	cl := redis.NewClient(&redis.Options{
		Addr: s.Addr().String(),
	})

	//  this library does not support Ping argument
	st, e := cl.Ping().Result()
	require.NoError(t, e)
	require.Equal(t, "PONG", st)

	require.NoError(t, s.ExpectationsWereMet())
}

func TestGoRedisList(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	v := map[string]string{"v1": "v2", "v3": "v4"}
	s.ExpectHGetAll("mykey", v).Once()

	cl := redis.NewClient(&redis.Options{
		Addr: s.Addr().String(),
	})

	ret, err := cl.HGetAll("mykey").Result()
	require.NoError(t, err)
	require.Equal(t, v, ret)

	require.NoError(t, s.ExpectationsWereMet())
}

func TestGoRedisListHSet(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, "")
	require.NoError(t, err)

	s.ExpectHSet("mykey", "fld", "value", true).Once()

	cl := redis.NewClient(&redis.Options{
		Addr: s.Addr().String(),
	})

	ret, err := cl.HSet("mykey", "fld", "value").Result()
	require.NoError(t, err)
	require.False(t, ret)

	require.NoError(t, s.ExpectationsWereMet())
}
