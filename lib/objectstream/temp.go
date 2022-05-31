package objectstream

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// 实现建立临时数据流; 实现提交函数(哈希值一致时候就提交,不一致就删除)
type TempPutStream struct {
	Server string
	Uuid   string
}

//建立临时数据流
// 这里就是在data节点上使用post将客户端put对象的size和散列值发送过去;
//这里的object对应的
func NewTempPutStream(server, hash string, size int64) (*TempPutStream, error) {
	request, e := http.NewRequest("POST", "http://"+server+"/temp/"+hash, nil)

	if e != nil {
		return nil, e
	}

	request.Header.Set("size", fmt.Sprintf("%d", size)) //请求头部增加size
	client := http.Client{}
	response, e := client.Do(request)
	if e != nil {
		return nil, e
	}
	//uuid相当于临时文件的文件名
	uuid, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return nil, e
	}
	return &TempPutStream{server, string(uuid)}, nil

}

func (w *TempPutStream) Commit(good bool) {
	method := "DELETE"
	if good {
		method = "PUT"
	}
	request, _ := http.NewRequest(method, "http://"+w.Server+"/temp/"+w.Uuid, nil)
	client := http.Client{}
	client.Do(request)

}

func NewTempGetStream(server, uuid string) (*GetStream, error) {
	return newGetStream("http://" + server + "/temp/" + uuid)
}

//write方法根据server和uuid属性的值,以patch方法访问数据的temp接口;将需要写入的数据上传
//在io.TeeReader(r, stream)中调用了该方法
func (w *TempPutStream) Write(p []byte) (n int, err error) {
	request, e := http.NewRequest("PATCH", "http://"+w.Server+"/temp/"+w.Uuid, strings.NewReader(string(p)))
	if e != nil {
		return 0, e
	}
	client := http.Client{}
	r, e := client.Do(request)
	if e != nil {
		return 0, e
	}
	if r.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("dataServer return http code %d", r.StatusCode)
	}
	return len(p), nil
}
