/**
 * @author zhougonghao
 * @date 2021/8/14
 * @note 缓存值的抽象与封装 (抽象的只读格式表示缓存值)
 */
package swift

// 缓存值表示结构体，只有一个字符数组属性
type ByteView struct {
	b []byte
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @return int 返回缓存值长度
 * @description 实现缓存值 lru.Value 接口
 **/
func (v ByteView) Len() int {
	return len(v.b)
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @return []byte
 * @description 返回缓存值的字节数组copy
 **/
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @return string
 * @description 获取缓存值字符串
 **/
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
