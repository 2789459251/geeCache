package geecache

import (
	"fmt"
	pb "geeCache/geecache/geecachepb"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
	"net/url"
)

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) //分布式节点
}

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error //一个分布式节点--组名+key值获取 缓存
}
type httpGetter struct { //httpGetter是一个分布式节点同时也是一个客户端
	baseURL string //将要访问的远程节点
}

var _PeerGetter = (*httpGetter)(nil) //接口类型变量赋值为nil

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	res, err := http.Get(u) //这就是路由的新请求
	if err != nil {
		return err //
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("拒绝服务：%v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("读取响应体：%v", err)
	}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("解码失败:%v", err)
	}
	return nil
} //客户端
