package put

import (
	"Storage/lib/es/ES8"
	"Storage/lib/utils"
	"log"
	"net/http"
	"strings"
)

// put的流程再梳理下;
// 收的put的请求后,实际还是选择一个随机data服务器,建立一个写入数据流(临时数据流);
// 等写入完毕计算该写入对象内容的哈希值是否与用户上传的一致,如果一致就将临时文件转为正式存储文件;
//如果不一致,则删除临时文件
func Put(w http.ResponseWriter, r *http.Request) {
	hash := utils.GetHashFromHeader(r.Header)
	if hash == "" {
		log.Println("missing object hash in digest header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	size := utils.GetSizeFromHeader(r.Header)
	c, e := storeObject(r.Body, hash, size)
	if e != nil {
		log.Println(e)
		w.WriteHeader(c)
		return
	}

	// put成功则在ES服务器中添加版本
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	e = ES8.AddVersion(name, size, hash)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
