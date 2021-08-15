/**
 * @author zhougonghao
 * @date 2021/8/14
 * @note
 */
package swift

//PeerPicker 节点选择接口
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
