# Redimock 


[![GoDoc](https://godoc.org/github.com/fzerorubigd/redimock?status.svg)](https://godoc.org/github.com/fzerorubigd/redimock)
[![Build Status](https://travis-ci.org/fzerorubigd/redimock.svg?branch=master)](https://travis-ci.org/fzerorubigd/redimock)
[![Coverage Status](https://coveralls.io/repos/github/fzerorubigd/redimock/badge.svg?branch=master)](https://coveralls.io/github/fzerorubigd/redimock?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/fzerorubigd/redimock)](https://goreportcard.com/report/github.com/fzerorubigd/redimock)

Redimock is the Redis mocking library in TCP level. This is not a Redis clone; it's just a mock. You need to know what command should be expected and provide the output for that commands. This information is available in Redis documents, but also I am adding more helper function to cover all Redis commands easily. 

## Usage

For usage, you can see the tests suits of this library itself. Currently, I do test it using redigo and go-redis. 
```go
package main 

import (
	"github.com/gomodule/redigo/redis"
)

func ReadRedis(red redis.Conn) error {
	v, err := redis.String(red.Do("GET", "KEY"))
	if err != nil {
		return err
	}

	_, err = red.Do("SET", v, "HI")
	if err != nil {
		return err
	}

	return nil
}
```

You can write test like this:

```go
package main 

import (
	"context"
	"testing"

	"github.com/fzerorubigd/redimock"
	"github.com/gomodule/redigo/redis"
)

func TestReadRedis(t *testing.T) {
	ctx, cl := context.WithCancel(context.Background())
	defer cl()

	mock, err := redimock.NewServer(ctx, "")
	if err != nil {
		t.FailNow()
	}

	rd, err := redis.Dial("tcp", mock.Addr().String())
	if err != nil {
		t.FailNow()
	}

	mock.ExpectGet("KEY", true, "ANOTHER")
	// Also it works with
	// mock.Expect("GET").WithArgs("KEY").WillReturn(redimock.BulkString("ANOTHER"))
	mock.Expect("SET").WithAnyArgs().WillReturn("OK")

	err = ReadRedis(rd)
	if err != nil {
		t.FailNow()
	}
}

```

The helper functions are not complete and all are subject to change. (functions inside the `commands.go` file)
