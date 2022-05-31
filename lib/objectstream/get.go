/*
实现apiserver与指定dataserver进行http通信,并完成get请求; 请求的正文是object对象的名称;
dataserver返回的http响应正文是object对象的具体内容;
*/

package objectstream

import (
	"fmt"
	"io"
	"net/http"
)

// GetStream 只用返回一个可读的reader即可;
type GetStream struct {
	reader io.Reader
}

func NewGetStream(server, objectName string) (*GetStream, error) {
	if server == "" || objectName == "" {
		return nil, fmt.Errorf("invalid server %s object %s", server, objectName)
	}

	return newGetStream("http://" + server + "/objects/" + objectName)
}

func newGetStream(url string) (*GetStream, error) {
	//与dataserver进行http通信,并获得dataserver的response响应
	r, e := http.Get(url)
	if e != nil {
		return nil, e
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %d", r.StatusCode)
	}

	return &GetStream{r.Body}, nil
}

func (r *GetStream) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}
