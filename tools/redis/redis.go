package redis

import (
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
	"time"
)

const (
	redisPort string = "127.0.0.1:6379"
	//redisPassword string = ""
)

var (
	con *redigo.Pool
)

func newPool() *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", redisPort)
			if err != nil {
				fmt.Println("Error :", err)
				return nil, err
			}
			// if _, err := c.Do("AUTH", password); err != nil {
			// 	c.Close()
			// 	return nil, err
			// }
			return c, nil
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func Get() redigo.Conn {
	con := newPool()
	return con.Get()
}

func Close() {
	con.Close()
}
