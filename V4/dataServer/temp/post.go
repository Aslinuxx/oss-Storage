package temp

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//post函数主要用于处理api服务器的post请求;
//在data服务器新建一个uuid文件 和uuid.dat文件;
//其中uuid文件中写入api服务传来的对象内容的散列值(hash值),对象内容的size;
//uuid.dat 在处理patch请求是会写入对象的内容; (即put对象的具体内容)
// 因为put对象的内容可能非常大所以该post中只写入uuid文件的内容;

type tempInfo struct {
	Uuid string
	Name string
	Size int64
}

func post(w http.ResponseWriter, r *http.Request) {
	//先生成一个uuid; uuid是类似hash值的作用; 生成一个唯一标识符
	output, _ := exec.Command("uuidgen").Output()
	uuid := strings.TrimSuffix(string(output), "\n")

	name := strings.Split(r.URL.EscapedPath(), "/")[2]       //实际上此时传递来的是hash值
	size, e := strconv.ParseInt(r.Header.Get("size"), 0, 64) // size
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//将api服务器传递来的hash值,size,以及生成的临时文件对应的uuid值全部存到结构中;并写入uuid文件中
	t := tempInfo{uuid, name, size}
	e = t.writeToFile() //写入uuid文件
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//新建uuid.dat文件
	_, e = os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid + ".dat")
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//将生成的uuid返回给api服务器
	w.Write([]byte(uuid))
}

func (t *tempInfo) writeToFile() error {
	f, e := os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid)
	if e != nil {
		return e
	}
	defer f.Close()
	b, _ := json.Marshal(t)
	f.Write(b)
	return nil
}
