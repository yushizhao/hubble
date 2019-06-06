package redis

import (
	"testing"
)

//测试get set
func Test0(t *testing.T) {
	client := NewClient("45.249.244.59:6666", "", 3, 60)
	defer client.OnClose()

	k := "name"
	v := "wudian"
	var body []byte

	if err := client.Insert(k, []byte(v)); err != nil {
		t.Error(err)
		return
	}

	ok, err := client.IsExist(k)
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Logf("%s  exist", k)
	}

	if body, err = client.Get(k); err != nil {
		t.Error(err)
		return
	} else {
		if string(body) == v {
			t.Logf("get %s is %s", k, v)
		} else {
			t.Errorf("want %s, but %s", v, string(body))
		}
	}

	if err := client.Del(k); err != nil {
		t.Error(err)
		return
	}
	ok, err = client.IsExist(k)
	if err != nil {
		t.Error(err)
		return
	}
	if !ok {
		t.Logf("%s not exist", k)
	}

	return
}

//测试hget hset
func Test1(t *testing.T) {
	client := NewClient("45.249.244.59:6666", "", 3, 60)
	defer client.OnClose()

	k := "hash"
	f1 := "name"
	v1 := "wudian"
	// f2 := "age"
	// v2 := 20
	// var body []byte

	if err := client.HInsert(k, f1, v1); err != nil {
		t.Error(err)
		return
	}

	//exist, _ := client.HIsExist(k, f1)
	//if exist {
	//	t.Log("exist")
	//} else {
	//	t.Log("not exist")
	//}
	//
	//time.Sleep(6 * time.Second)

	exist, _ := client.HIsExist(k, f1)
	if exist {
		t.Log("exist")
	} else {
		t.Log("not exist")
	}

	if body, err := client.HGet(k, f1); err != nil {
		t.Error(err)
		return
	} else {
		t.Logf(string(body))
	}

	if err := client.HDel(k, f1); err != nil {
		t.Error(err)
		return
	}
	_, err := client.IsExist(k, f1)
	if err != nil {
		t.Error(err)
		return
	}

	// return
}

//测试keys *
func Test2(t *testing.T) {
	client := NewClient("45.249.244.59:6666", 3, 60)
	defer client.OnClose()

	val, err := client.Keys("DEPTHx|EOS_BTC.*")
	if err == nil {
		for i, v := range val {
			t.Log(i, v)
		}
	}
}
