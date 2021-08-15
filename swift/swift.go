/**
 * @author zhougonghao
 * @date 2021/8/14
 * @note
 */
package swift

import (
	"fmt"
	"github.com/swift-cache/swift/singleflight"
	"log"
	"sync"
)

/************************************Getter定义***************************************************/

//定义接口 Getter 和 回调函数 Get(key string)([]byte, error)，参数是 key，返回值是 []byte。
//定义函数类型 GetterFunc，并实现 Getter 接口的 Get 方法。
//函数类型实现某一个接口，称之为接口型函数，方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数。

//回调 Getter 接口，用户实现该接口，当缓存不存在时，调用该函数加载源数据
type Getter interface {
	Get(key string) ([]byte, error)
}

//声明 Getter 接口的实现
type GetterFunc func(key string) ([]byte, error)

//实现 Getter 接口
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

/************************************Group定义 实现单机并发***************************************************/
//Group 定义缓存组
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @param  name 组名称
 * @param  cacheBytes 缓存大小
 * @param  getter 回调
 * @return *Group 组指针
 * @description 缓存组构造方法
 **/
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mutex.Lock()
	defer mutex.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @param  name 组名称
 * @return Group 组指针
 * @description 通过组名获取组
 **/
func GetGroup(name string) *Group {
	mutex.RLock()
	g := groups[name]
	mutex.RUnlock()
	return g
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @param  key 缓存键
 * @return ByteView
 * @description 实现缓存组的 Get 方法
 **/
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocal(key)
	})
	if err == nil {
		return view.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @param
 * @return
 * @description getLocal 调用用户回调函数
 **/
func (g *Group) getLocal(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	val := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, val)
	return val, nil
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @description g.getter.Get() 获取源数据，并且将源数据添加到缓存 mainCache 中
 **/
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @description 节点注册
 **/
// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}
