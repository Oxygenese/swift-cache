/**
 * @author zhougonghao
 * @date 2021/8/14
 * @note
 */
package swift

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

/**
 * @date   2021/8/14
 * @author zhougonghao
 * @description 测试回调函数接口是否生效
 **/
func TestGetterFunc_Get(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expected := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expected) {
		t.Errorf("callback failed!")
	}
}

/****************************单机并发测试******************************************/

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

//创建 group 实例，并测试 Get 方法
//在这个测试用例中，我们主要测试了 2 种情况
//1）在缓存为空的情况下，能够通过回调函数获取到源数据。
//2）在缓存已经存在的情况下，是否直接从缓存中获取，为了实现这一点，使用 loadCounts 统计某个键调用回调函数的次数，如果次数大于1，则表示调用了多次回调函数，没有缓存。
func TestGroup_Get(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	swift := NewGroup("scores", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDb] search key", key)
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key] += 1
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
	for k, v := range db {
		if view, err := swift.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		} // load from callback function
		if _, err := swift.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := swift.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
