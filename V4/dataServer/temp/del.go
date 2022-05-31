package temp

import (
	"net/http"
	"os"
	"strings"
)

//对commit(false)的处理
//删除两个新建的文件见uuid 和 uuid.dat
func del(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"
	os.Remove(infoFile)
	os.Remove(datFile)
}
