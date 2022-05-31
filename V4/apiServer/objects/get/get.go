package get

import (
	"Storage/lib/es/ES8"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func Get(w http.ResponseWriter, r *http.Request) {

	//先拿到get的名称和版本
	objectName := strings.Split(r.URL.EscapedPath(), "/")[2]

	versionArray := r.URL.Query()["version"]
	version := 0
	var err error
	if len(versionArray) != 0 {
		version, err = strconv.Atoi(versionArray[0])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// 根据name version 去获取对应的hash值; 然后建立一个文件流,以该hash值为查询,获取对应data服务器上的对象具体数据内容
	meatdata, err := ES8.GetMetadata(objectName, version)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if meatdata.Hash == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hash := url.PathEscape(meatdata.Hash)
	stream, err := getStream(hash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	io.Copy(w, stream)
}
