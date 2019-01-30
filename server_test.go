package server

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewServer(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	s, err := NewServer(ctx, ":12345")
	require.NoError(t, err)

	s.ExpectGet("test_key", true, "Hello").Once()

	cl := redis.NewClient(&redis.Options{
		Addr: s.Addr().String(),
	})

	str := cl.Get("test_key")
	x, e := str.Result()
	require.NoError(t, e)
	require.Equal(t, "Hello", x)

	require.NoError(t, s.ExpectationsWereMet())
}
