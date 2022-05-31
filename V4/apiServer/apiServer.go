package main

import (
	"Storage/V4/apiServer/heartbeat"
	"Storage/V4/apiServer/locate"
	"Storage/V4/apiServer/objects"
	"Storage/V4/apiServer/version"
	"log"
	"net/http"
	"os"
)

func main() {
	//监听data服务器的心跳信息
	go heartbeat.ListenHeartbeat()

	//put/get/del请求
	http.HandleFunc("/objects/", objects.Handler)

	//locate请求; 这里locate是按照散列值进行locate
	http.HandleFunc("/locate/", locate.Handler)

	//实现version请求
	http.HandleFunc("/version/", version.Handler)

	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))

}
