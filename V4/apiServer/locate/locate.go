package locate

import (
	"Storage/lib/rabbitmq"
	"os"
	"strconv"
	"time"
)

//每当需要定位时候,api服务器都会建立一个临时消息队列,向dataServer exchange中广播定位消息;
//消息的正文即是需要定位的object的名称; 如果定位成功就会收到反馈,超过一定时间没有反馈则认为定位失败
func Locate(objectName string) string {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))

	q.Publish("dataServers", objectName)
	//广播完再开始等回应,所以channel后建立
	c := q.Consume()
	go func() {
		time.Sleep(time.Second)
		q.Close()

	}()

	msg := <-c
	s, _ := strconv.Unquote(string(msg.Body)) // 去除引号

	return s

}

func Exist(objectName string) bool {
	return Locate(objectName) != ""
}
