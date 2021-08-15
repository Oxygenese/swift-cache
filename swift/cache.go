/**
 * @author zhougonghao
 * @date 2021/8/14
 * @note 并发控制
 * 对lru进行并发控制封装(线程安全)
 * cache.go 的实现非常简单，实例化 lru，封装 get 和 add 方法，并添加互斥锁 mu。
 * 在 add 方法中，判断了 c.lru 是否为 nil，如果等于 nil 再创建实例。
 * 这种方法称之为延迟初始化(Lazy Initialization)，一个对象的延迟初始化意味着该
 * 对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求。
 */
package swift

import (
	"github.com/swift-cache/swift/lru"
	"sync"
)

type cache struct {
	//互斥锁
	mu sync.Mutex
	//lru数据结构
	lru *lru.Cache
	//缓存大小
	cacheBytes int64
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @param  key 缓存键
 * @param  value 缓存值
 * @description 添加缓存值
 **/
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @param  key 缓存键
 * @return val ByteView,ok bool
 * @description
 **/
func (c *cache) get(key string) (val ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
