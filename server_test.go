package server

import (
	"context"
	"testing"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx)
	require.NoError(t, err)

	s.ExpectGet("test_key", true, "Hello").Once()
	s.ExpectGet("test_key_2", false, "").Once()

	cl := redis.NewClient(&redis.Options{
		Addr: s.Addr().String(),
	})

	str := cl.Get("test_key")
	x, e := str.Result()
	require.NoError(t, e)
	require.Equal(t, "Hello", x)

	str = cl.Get("test_key_2")
	_, e = str.Result()
	require.Error(t, e)

	require.NoError(t, s.ExpectationsWereMet())
}

func TestServer_Close(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx)
	require.NoError(t, err)

	s.ExpectGet("abcd", true, "test").Once()

	require.Error(t, s.ExpectationsWereMet())
}
