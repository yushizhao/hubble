package redis

import "github.com/mediocregopher/radix"

func Get(client radix.Client, key string) ([]byte, error) {
	var b []byte
	err := client.Do(radix.Cmd(&b, "GET", key))
	return b, err
}

func HGet(client radix.Client, key, field string) ([]byte, error) {
	var b []byte
	err := client.Do(radix.Cmd(&b, "HGET", key, field))
	return b, err
}

func HKeys(client radix.Client, key string) ([]string, error) {
	var s []string
	err := client.Do(radix.Cmd(&s, "HKEYS", key))
	return s, err
}

func Keys(client radix.Client, key string) ([]string, error) {
	var s []string
	err := client.Do(radix.Cmd(&s, "KEYS", key))
	return s, err
}
