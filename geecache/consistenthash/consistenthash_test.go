package consistenthash

import (
	"strconv"
	"testing"
)

// 测试通过证明 添加节点只影响了该节点附近的少部分数据
func TestHashing(t *testing.T) {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4", //!!!
		"27": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("请求k:%s,应该位于v :%s", k, v)
		}
	}

	hash.Add("8")
	testCases["27"] = "8"
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("请求k:%s,应该位于v :%s", k, v)
		}
	}

}
