package lru

import (
	"reflect"
	"testing"
)

type String string

func (b String) Len() int {
	return len(b)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("1234"))

	if v, ok := lru.Get("key1"); !ok || v.(String) != "1234" {
		t.Fatalf("缓存击中 key1==1234 失败")
	}
}
func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("12345"))
	lru.Add("k2", String("v2"))
	lru.Add("k3", String("v3"))
	lru.Add("k4", String("v4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("回调函数错误")
	}
}
