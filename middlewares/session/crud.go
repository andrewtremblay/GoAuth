package session

import (
	"authsys/tools/context"
	"authsys/tools/redis"
	redigo "github.com/garyburd/redigo/redis"
	"net/http"
	"time"
)

/*
 * Session data processor
 */

// Need request object, for reading session id from request context
func InsertData(r *http.Request, key string, value interface{}) {
	sid := context.Get(r, context.SID).(string)
	c := redis.Get()
	if _, err := c.Do("HMSET", sid, key, value); err != nil {
		panic(err.Error())
	}

	// Session storage id will be remove in 10 hours. Keep
	// redis database clean.
	c.Do("EXPIREAT", sid, time.Now().Add(time.Hour*10).Unix())
}

// If value is not found, then return empty string
func ReadData(r *http.Request, key string) (interface{}, error) {

	var value string

	sid := context.Get(r, context.SID).(string)
	c := redis.Get()
	raw, err := redigo.Values(c.Do("HMGET", sid, key))

	redigo.Scan(raw, &value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func DeleteData(r *http.Request, key string) {
	sid := context.Get(r, context.SID).(string)
	c := redis.Get()
	c.Do("HDEL", sid, key)
}
