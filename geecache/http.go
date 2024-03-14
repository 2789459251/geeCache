package geecache

import (
	"fmt"
	"geeCache/geecache/consistenthash"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

var _PeerPicker = (*HTTPPool)(nil)

type HTTPPool struct {
	self        string //主机和端口
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter //客户端baseurl->分布式节点客户端
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil) //哈希环及其对应的节点与虚拟节点
	p.peers.Add(peers...)                              //添加虚拟节点与真实节点
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath} //将真实节点名称与客户端映射
	}
}
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.peers.Get(key); peer != "" && peer != p.self { //peer :真实节点名称
		p.Log("匹配远程节点：%s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}
func (p *HTTPPool) Log(farmat string, v ...interface{}) {
	log.Printf("[server %s] %s", p.self, fmt.Sprintf(farmat, v...))
}
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool接收到非缓存类的请求：" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/obtet-stream")
	w.Write(view.ByteSlice())
}
