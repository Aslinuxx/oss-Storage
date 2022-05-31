package rabbitmq

import (
	"encoding/json"
	"github.com/streadway/amqp" // github上对rabbitmq进行封装好,简化使用的包
)

type RabbitMQ struct{
	channel *amqp.Channel // 消息队列
	conn 	*amqp.Connection
	Name     string
	exchange	string // 首字母小写表示该成员是私有的不能被外部访问;只能包内使用

}

// new一个rabbitmq结构体
func New(s string) *RabbitMQ{
	conn, e := amqp.Dial(s)
	if e != nil{
		panic(e)
	}

	ch,e := conn.Channel()
	if e != nil{
		panic(e)
	}

	// 消息队列
	q,e := ch.QueueDeclare(
		"",//name
		false,//durbale
		true,//delete when unused
		false,// exclusive
		false,// no-wait
		nil,//arguments
		)

	if e != nil {
		panic(e)
	}

	mq := new(RabbitMQ)
	mq.channel = ch
	mq.conn = conn
	mq.Name = q.Name
	return mq
}

// 将消息队列与exchange(消息处理中心)进行绑定
func (q *RabbitMQ) Bind(exchange string){
	e := q.channel.QueueBind(
		q.Name,//queue name 队列名称
		"",//routing key
		exchange,//exchange
		false,//no-wait
		nil,//arguments
		)

	if e != nil{
		panic(e)
	}
	q.exchange = exchange
}

//这是向某一队列发送消息;
//即一对一通信; A队列向B队列发送信息
func (q *RabbitMQ) Send(queue string, body interface{}){
	str, e := json.Marshal(body) // json序列化; 将body的格式转为json格式
	if e != nil {
		panic(e)
	}

	e = q.channel.Publish(
		"", // exchange
		queue,// B队列名称;
		false,//mandatory
		false,//immediate

		//这是该队列收到消息后,进行反馈的队列名称(即A队列), 已经需要反馈的消息内容
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    []byte(str),
		})
	if e != nil{
		panic(e)
	}
}

//通过exchange广播到所有连接到该exchange的队列
func (q *RabbitMQ) Publish(exchange string, body interface{}){
	str, e := json.Marshal(body)
	if e != nil{
		panic(e)
	}

	e = q.channel.Publish(
		exchange,//exchange
		"",//一对一通信队列名称
		false,
		false,
		amqp.Publishing{
			ReplyTo:  q.Name,
			Body: []byte(str),
		})
	if e != nil{
		panic(e)
	}


}

//Consume()是用于生成一个接收消息的 channel 通道;
//比如此时api服务器的消息队列都绑定在apiServers exchange上了;
// 那么该消息队列收到的信息就用该通道来传递; 实际上此时该通道内就是data服务器的心跳信息;
//(还在活跃的data服务器每5秒向apiserver exchange发送一次心跳信息;由该exchange广播给所有连接到此的消息队列)
func (q *RabbitMQ) Consume() <-chan amqp.Delivery{
	c, e := q.channel.Consume(
		q.Name,// queue 队列名称
		"", // consumer
		true,
		false,
		false,
		false,
		nil,
		)
	if e != nil{
		panic(e)
	}

	return c
}

func (q *RabbitMQ) Close() {
	q.channel.Close()
	q.conn.Close()
}