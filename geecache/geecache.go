package geecache

import (
	"fmt"
	pb "geeCache/geecache/geecachepb"
	"geeCache/geecache/singleflight"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

/*不太理解
Getter 是一个接口，要去实现get方法
GetterFunc 是一个函数（类型），参数返回与get方法相同
GetterFunc实现get方法，实质上是调用了自己
*/

type Group struct {
	name      string
	getter    Getter //缓存未命中，获取值的方法
	mainCache cache
	peers     PeerPicker //服务端

	loader *singleflight.Group
}

var (
	mu     sync.RWMutex              //读锁，大家都可以获取读锁，但是写入需要抢写锁，写锁被抢占后，其他协程就不能获取读锁
	groups = make(map[string]*Group) //分布式缓存的体现
)

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("注册的服务端已超过一台")
	}
	g.peers = peers
}
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter") // 数据源转换成字节类型
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}
func GetGroup(name string) *Group {
	mu.RLock() //用于获取读取锁。多个 goroutine 可以同时获取读锁，只有当有 goroutine 持有写锁时，调用该方法的 goroutine 才会被阻塞
	g := groups[name]
	mu.RUnlock()
	return g
}
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[geecache]hit") //断言value是ByteView类型
		return v, nil
	}
	return g.load(key)
}
func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) { //只请求一次
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, err
				}
				log.Println("从远程节点中获取缓存失败", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 使用回调函数，获取处理过的数据源
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {

	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}

	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
