package redis

import (
	//"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Start the redis server, before run this test
func TestConnectionToRedis(t *testing.T) {
	Convey("Redis connection test.", t, func() {

		conn := Get()
		_, err := conn.Do("HMSET", "id", "Key", "Value")
		assert.Nil(t, err)
	})
}
