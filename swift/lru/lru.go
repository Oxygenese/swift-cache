/**
 * @date   2021/8/14
 * @author zhougonghao
 * @description  lru 缓存淘汰策略
 **/
package lru

import "container/list"

// Cache 线程不安全的缓存结构体
type Cache struct {
	//最大内存
	maxBytes int64
	//剩余内存
	nbytes int64
	//双向链表
	ll *list.List
	//字典，存储键值映射关系
	cache map[string]*list.Element
	// 可选回调函数，在元素移除时触发
	OnEvicted func(key string, value Value)
}
type entry struct {
	key   string
	value Value
}

// Value 节点值接口
type Value interface {
	Len() int
}


/**
 * @date   2021/8/14
 * @author zhougonghao
 * @param  maxBytes 最大内存大小
 * @param  onEvicted (可选)移除元素时的回调函数
 * @return *Cache
 * @description
 **/
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @param  key 键
 * @return value Value, ok bool
 * @description 查找主要有 2 个步骤，第一步是从字典中找到对应的双向链表的节点，第二步，将该节点移动到队尾。
 *  c.ll.MoveToFront(ele)，即将链表中的节点 ele 移动到队尾（双向链表作为队列，队首队尾是相对的，在这里约定 front 为队尾）
 **/
func (c *Cache) Get(key string) (value Value, ok bool) {
	//根据键查找元素
	if ele, ok := c.cache[key]; ok {
		//若存在则将其放在队首
		c.ll.MoveToFront(ele)
		//通过断言将value转为entry对象
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}
/**
 * @date   2021/8/14
 * @author zhougonghao
 * @description 这里的删除，实际上是缓存淘汰。即移除最近最少访问的节点（队首）
 * c.ll.Back() 取到队首节点，从链表中删除。
 * delete(c.cache, kv.key)，从字典中 c.cache 删除该节点的映射关系。
 * 更新当前所用的内存 c.nbytes。
 * 如果回调函数 OnEvicted 不为 nil，则调用回调函数。
 **/
// RemoveOldest
func (c *Cache) RemoveOldest() {
	//取出链表最后一个元素
	ele := c.ll.Back()
	if ele != nil {
		//若最后一个元素不为空,就移除
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @param  key 键
 * @param  value 值
 * @description 新增/修改
 * 如果键存在，则更新对应节点的值，并将该节点移到队尾。
 * 不存在则是新增场景，首先队尾添加新节点 &entry{key, value}, 并字典中添加 key 和节点的映射关系。
 * 更新 c.nbytes，如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点。
 **/
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len 最后，为了方便测试，我们实现 Len() 用来获取添加了多少条数据。
func (c *Cache) Len() int {
	return c.ll.Len()
}
