package objects

import (
	"net/http"
	"strings"
)

//这里在处理get请求之前;必须加一步数据校验;
//即在将data服务器上的存的对应的对象文件发回给api服务器时,也必须做一步校验;
//因为即使数据原封不动存在data节点上,也可能会发生随着时间的流逝原来存储的正确数据此时也可能是错误的;
//即此时的虽然存储的文件的内容可能发生了损坏,所以仍需进行一步校验其内容的hash值是否还是原来数据的hash值(即文件名)
func get(w http.ResponseWriter, r *http.Request) {
	file := getFile(strings.Split(r.URL.EscapedPath(), "/")[2])
	if file == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 如果没问题就正常将该文件的内容返回给api服务器
	sendFile(w, file)
}
