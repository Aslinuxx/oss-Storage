package locate

import (
	"Storage/lib/rabbitmq"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var objects = make(map[string]int)
var mutex sync.Mutex

func CollectObjects() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*") // *表示该目录下的所有文件
	for i := range files {
		hash := filepath.Base(files[i])
		objects[hash] = 1
	}

}

func Locate(hash string) bool {
	mutex.Lock()
	_, ok := objects[hash]
	mutex.Unlock()
	return ok
}

//data服务节点将自己的消息队列绑定在16:35 2022/5/14 exchange上; 每当收到locate消息时候就需要对其进行响应;

func StartLocate() {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()

	q.Bind("dataServers")

	c := q.Consume()

	for msg := range c {
		hash, e := strconv.Unquote(string(msg.Body))
		if e != nil {
			panic(e)
		}

		exist := Locate(hash)
		if exist {
			q.Send(msg.ReplyTo, os.Getenv("LISTEN_ADDRESS"))
		}
	}

}

func Add(hash string) {
	mutex.Lock()
	objects[hash] = 1
	mutex.Unlock()
}

func Del(hash string) {
	mutex.Lock()
	delete(objects, hash)
	mutex.Unlock()
}
