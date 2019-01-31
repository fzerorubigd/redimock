package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoRedisServerExpectationNotMet(t *testing.T) {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	s, err := NewServer(ctx, ":12345")
	require.NoError(t, err)

	s.ExpectGet("abcd", true, "test").Once()

	require.Error(t, s.ExpectationsWereMet())
}
