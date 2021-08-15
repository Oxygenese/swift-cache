/**
 * @author zhougonghao
 * @date 2021/8/14
 * @note
 */
package main

import (
	"fmt"
	"github.com/swift-cache/swift"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

//$ curl http://localhost:9999/_geecache/scores/Tom
//630
//$ curl http://localhost:9999/_geecache/scores/kkk
//kkk not exist
func main() {
	swift.NewGroup("scores", 2<<10, swift.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := ":8000"
	peers := swift.NewHTTPPool(addr)
	log.Println("swift is running at", addr)
	err := http.ListenAndServe(addr, peers)
	if err != nil {
		log.Println(err)
		return
	}
}
