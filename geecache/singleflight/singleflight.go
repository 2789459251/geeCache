package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup //  仅执行一次，防止锁重入
	val interface{}
	err error
}
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call) //延迟初始化
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() //如果请求正在进行，则等待
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1) //锁加1
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done() //锁减1

	g.mu.Lock()
	delete(g.m, key) //这是什么更新原理
	g.mu.Unlock()

	return c.val, c.err

}
