package server

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
