package del

import (
	"Storage/lib/es/ES8"
	"log"
	"net/http"
	"strings"
)

//实现对对象的最新版本的删除;
//逻辑删除
func Del(w http.ResponseWriter, r *http.Request) {
	//只需实现逻辑删除;
	//即取ES中找出该对象的最新版本; 然后再ES中再插入一条最新版本,将size, hash全部变为0即可

	objectName := strings.Split(r.URL.EscapedPath(), "/")[2]

	version, e := ES8.SearchLatestVersion(objectName)

	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//插入一条空版本作为最新版本即代表删除
	//e = ES8.AddVersion(objectName,0,"")
	//注意 这里还是用插入一条元数据更好; 因为如果原数据版本已经被删除了;
	e = ES8.PutMetadata(objectName, version.Version+1, 0, "")
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
