package geeCache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	}) //强转

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatalf("回调失败")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "598",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[showDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	for k, v := range db {
		//v是[]byte吧
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("无法获得Tom的值")
		}
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("缓存%s失踪", k)
		}
	}
	if view, err := gee.Get("unknow"); err == nil {
		t.Fatalf("unknow的值应该为空，但是获取到了%s", view)
	}
	gee.Get("Tom")
}
