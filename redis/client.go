package redis

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

func (c *Client) Get(args ...interface{}) (value []byte, err error) {

	conn := c.pool.Get()
	defer conn.Close()

	return r.Bytes(conn.Do("GET", args[0].(string)))

}

func (c *Client) IsExist(args ...interface{}) (ok bool, err error) {
	conn := c.pool.Get()
	defer conn.Close()
	ok, err = r.Bool(conn.Do("EXISTS", args[0].(string)))
	if err != nil {
		return
	}

	return
}

func (c *Client) Del(args ...interface{}) (err error) {

	_, err = c.pool.Get().Do("DEL", args[0].(string))

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
func (c *Client) HGet(args ...interface{}) (value []byte, err error) {

	conn := c.pool.Get()
	defer conn.Close()

	return r.Bytes(conn.Do("HGET", args[0].(string), args[1].(string)))

}

func (c *Client) HGetAll(args ...interface{}) (value []byte, err error) {

	conn := c.pool.Get()
	defer conn.Close()
	raw, err := r.StringMap(conn.Do("HGETALL", args[0].(string)))
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(raw)

	return b, err
}

func (c *Client) HKeys(args ...interface{}) (value []string, err error) {

	conn := c.pool.Get()
	defer conn.Close()
	raw, err := r.Strings(conn.Do("HKEYS", args[0].(string)))

	return raw, err
}

func (c *Client) HIsExist(args ...interface{}) (ok bool, err error) {
	conn := c.pool.Get()
	defer conn.Close()
	ok, err = r.Bool(conn.Do("HEXISTS", args[0].(string), args[1].(string)))
	if err != nil {
		return
	}

	return
}

func (c *Client) HDel(args ...interface{}) (err error) {

	_, err = c.pool.Get().Do("HDEL", args[0].(string), args[1].(string))

	return
}

func (c *Client) Keys(args ...interface{}) (value []string, err error) {

	conn := c.pool.Get()
	defer conn.Close()

	return r.Strings(conn.Do("KEYS", args[0].(string)))

}

func (c *Client) PSub(args ...interface{}) (r.PubSubConn, error) {
	psc := r.PubSubConn{Conn: c.pool.Get()}
	err := psc.PSubscribe(args[0].(string))
	return psc, err
}
