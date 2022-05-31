package heartbeat

import (
	"Storage/lib/rabbitmq"
	"os"
	"strconv"
	"sync"
	"time"
)

var dataServers = make(map[string]time.Time)
var mutex sync.Mutex

func ListenHeartbeat() {
	//api服务器的消息队列一直在监听来自data服务器的心跳信息;
	//data服务广播(publish)的是其自身的ip地址
	//其消息队列绑定到apiServers exchange上; 其是消费者;
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))

	defer q.Close()

	q.Bind("apiServers")

	//消息队列的channel
	c := q.Consume()

	go removeExpiredDataServer()

	for msg := range c {
		dataServer, err := strconv.Unquote(string(msg.Body))
		if err != nil {
			panic(err)
		}

		mutex.Lock()
		dataServers[dataServer] = time.Now()
		mutex.Unlock()

	}

}

func removeExpiredDataServer() {
	//每五秒清除一次过期的信息
	for {
		time.Sleep(5 * time.Second)

		mutex.Lock()
		for s, t := range dataServers {
			if t.Add(10 * time.Second).Before(time.Now()) {
				delete(dataServers, s)
			}
		}
		mutex.Unlock()
	}

}
