package geecache

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) //分布式节点
}

type PeerGetter interface {
	Get(group string, key string) ([]byte, error) //一个分布式节点--组名+key值获取 缓存
}
type httpGetter struct { //httpGetter是一个分布式节点同时也是一个客户端
	baseURL string //将要访问的远程节点
}

var _PeerGetter = (*httpGetter)(nil) //接口类型变量赋值为nil

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	res, err := http.Get(u)
	if err != nil {
		return nil, err //
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("拒绝服务：%v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体：%v", err)
	}
	return bytes, nil
} //客户端
