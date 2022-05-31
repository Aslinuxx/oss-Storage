package objects

import (
	"Storage/V4/apiServer/objects/del"
	"Storage/V4/apiServer/objects/get"
	"Storage/V4/apiServer/objects/put"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method

	if m == http.MethodGet {
		get.Get(w, r)
		return
	}

	if m == http.MethodDelete {
		del.Del(w, r)
		return
	}

	if m == http.MethodPut {
		put.Put(w, r)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
