package main

import (
	"Storage/V4/dataServer/heartbeat"
	"Storage/V4/dataServer/locate"
	"Storage/V4/dataServer/objects"
	"Storage/V4/dataServer/temp"
	"net/http"
	"os"
	"log"
)

func main() {
	//因为每次api服务器进行locate的时候,所有的data服务器都需要对该locate消息队列进行响应,查看locate的路径是否存在在自己的机子上;
	//使用os.Stat()进行判断; 这样的含义意味着每次locate请求所有机子都得进行一次磁盘访问;  这是非常频繁的IO; 会给系统带来很大的负担
	//所以此次改进成只有程序启动时候进行一次全机目录的扫描,然后将其存入内存中; 以k-v map的形式存入;
	//之后需要locate的时候就查询该map即可; 当然随着新的object的加入,该map也在更新
	locate.CollectObjects()

	go heartbeat.StartHeartbeat()
	go locate.StartLocate()

	http.HandleFunc("/temp/", temp.Handler)
	http.HandleFunc("/objects/", objects.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))

}
