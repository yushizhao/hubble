package rediswrapper

import (
	"encoding/json"
	"fmt"
	"time"

	r "github.com/gomodule/redigo/redis"
)

type Client struct {
	pool *r.Pool
}

func NewClient(ipport string, password string, maxidle int, timeout int) *Client {
	return &Client{
		pool: &r.Pool{
			MaxIdle:     maxidle,
			IdleTimeout: time.Duration(timeout) * time.Second,
			Dial: func() (r.Conn, error) {
				server := ipport
				c, err := r.Dial("tcp", server)
				if err != nil {
					return nil, err
				}
				if true {
					if _, err := c.Do("AUTH", password); err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, err
			},
		},
	}
}

func (c *Client) OnClose() {
	c.pool.Close()
}

func (c *Client) Insert(args ...interface{}) (err error) {
	conn := c.pool.Get()
	defer conn.Close()

	if len(args) < 2 {
		err = fmt.Errorf("Redis Insert Params nums Error")
		return
	}
	_, err = conn.Do("SET", args[0].(string), args[1])
	if err != nil {
		return
	}

	if len(args) > 2 {
		conn.Do("EXPIRE", args[0].(string), args[2].(int))
	}

	return
}

func (c *Client) Get(key string) (value []byte, err error) {

	conn := c.pool.Get()
	defer conn.Close()

	return r.Bytes(conn.Do("GET", key))

}

func (c *Client) IsExist(key string) (ok bool, err error) {
	conn := c.pool.Get()
	defer conn.Close()
	ok, err = r.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return
	}

	return
}

func (c *Client) Del(key string) (err error) {

	_, err = c.pool.Get().Do("DEL", key)

	return
}

func (c *Client) HInsert(args ...interface{}) (err error) {
	conn := c.pool.Get()
	defer conn.Close()

	if len(args) < 3 {
		err = fmt.Errorf("Redis HInsert Params nums Error")
		return
	}
	_, err = conn.Do("HSET", args[0].(string), args[1].(string), args[2])
	if err != nil {
		return
	}

	if len(args) > 3 {
		conn.Do("EXPIRE", args[0].(string), args[3].(int))
	}

	return
}
func (c *Client) HGet(key, field string) (value []byte, err error) {

	conn := c.pool.Get()
	defer conn.Close()

	return r.Bytes(conn.Do("HGET", key, field))

}

func (c *Client) HGetAll(key string) (value []byte, err error) {

	conn := c.pool.Get()
	defer conn.Close()
	raw, err := r.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(raw)

	return b, err
}

func (c *Client) HKeys(key string) (value []string, err error) {

	conn := c.pool.Get()
	defer conn.Close()
	raw, err := r.Strings(conn.Do("HKEYS", key))

	return raw, err
}

func (c *Client) HIsExist(key, field string) (ok bool, err error) {
	conn := c.pool.Get()
	defer conn.Close()
	ok, err = r.Bool(conn.Do("HEXISTS", key, field))
	if err != nil {
		return
	}

	return
}

func (c *Client) HDel(key, field string) (err error) {

	_, err = c.pool.Get().Do("HDEL", key, field)

	return
}

func (c *Client) Keys(pattern string) (value []string, err error) {

	conn := c.pool.Get()
	defer conn.Close()

	return r.Strings(conn.Do("KEYS", pattern))

}

func (c *Client) PSub(pattern string) (r.PubSubConn, error) {
	psc := r.PubSubConn{Conn: c.pool.Get()}
	err := psc.PSubscribe(pattern)
	return psc, err
}

func (c *Client) Sub(channel ...string) (r.PubSubConn, error) {
	psc := r.PubSubConn{Conn: c.pool.Get()}
	err := psc.Subscribe(r.Args{}.AddFlat(channel)...)
	return psc, err
}

func (c *Client) Pub(channel, msg string) error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("PUBLISH", channel, msg)
	return err
}
