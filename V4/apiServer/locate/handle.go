package locate

import (
	"encoding/json"
	"net/http"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {

	//取出request请求中的objectName; 然后进行访问;
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	server := Locate(strings.Split(r.URL.EscapedPath(), "/")[2])
	if server == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b, _ := json.Marshal(server)
	w.Write(b)

}
