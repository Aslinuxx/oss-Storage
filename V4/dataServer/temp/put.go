package temp

import (
	"Storage/V4/dataServer/locate"
	"log"
	"net/http"
	"os"
	"strings"
)

//根据api服务器上散列值的校验结果;进行处理
//这里是Commit(true);即校验一致,进行将临时文件转正;
func put(w http.ResponseWriter, r *http.Request) {
	//下面只是再检查一遍临时文件是否有错
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempinfo, e := readFromFile(uuid)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"
	f, e := os.Open(datFile)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	actual := info.Size()
	os.Remove(infoFile)
	if actual != tempinfo.Size {
		os.Remove(datFile)
		log.Println("actual size mismatch, expect", tempinfo.Size, "actual", actual)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//以上检查都没错,则正式提交
	commitTempobject(datFile, tempinfo)

}

//正式提交分为两步:1.将临时文件重命名为正式文件 2:将增加的文件加到开机扫描的文件内存找中去; map
func commitTempobject(datFile string, tempinfo *tempInfo) {
	os.Rename(datFile, os.Getenv("STORAGE_ROOT")+"/objects/"+tempinfo.Name) //此时name是hash值
	locate.Add(tempinfo.Name)
}
