package version

import (
	"Storage/lib/es/ES8"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

//返回指定对象的所有版本; 没有指定对象默认返回所有对象的所有版本;

func Handler(w http.ResponseWriter, r *http.Request) {

	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//取出http请求中的对象的objectName

	objectName := strings.Split(r.URL.EscapedPath(), "/")[2]

	form := 0
	size := 1000

	for {
		metadatas, err := ES8.SearchAllVersions(objectName, form, size)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for i := range metadatas {
			b, _ := json.Marshal(metadatas[i])
			w.Write(b)
			w.Write([]byte("\n"))

		}
		if len(metadatas) != size { // 不等于说明此时已经搜索到最后一页了; 页面的行数据已经不足1000了
			return
		}
		form += size

	}
}
