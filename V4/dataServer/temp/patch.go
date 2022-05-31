package temp

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

//实现对api服务器patch请求的处理;
//将api服务器传来的客户端put对象的具体内容写入uuid.dat
func patch(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempinfo, e := readFromFile(uuid)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"

	f, e := os.OpenFile(datFile, os.O_WRONLY|os.O_APPEND, 0) // 这里设置权限了; 因为datFile只可写入
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer f.Close()

	_, e = io.Copy(f, r.Body)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	info, e := f.Stat()
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	actual := info.Size()
	if actual > tempinfo.Size {
		os.Remove(datFile)
		os.Remove(infoFile)
		log.Println("actual size", actual, "exceeds", tempinfo.Size)
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func readFromFile(uuid string) (*tempInfo, error) {
	f, e := os.Open(os.Getenv("STORAGE_ROOT") + "/temp/" + uuid)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	var info tempInfo
	json.Unmarshal(b, &info)
	return &info, nil
}
