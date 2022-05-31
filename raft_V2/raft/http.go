package raft

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
)

type Message struct {
	MsgID   int
	MsgBody string
}

func (rf *Raft) httpListen() {
	http.HandleFunc("/req", rf.getRequest)
	fmt.Println("监听8090端口")
	if err := http.ListenAndServe(":8090", nil); err != nil {
		fmt.Println(err)
		return
	}
}

func (rf *Raft) getRequest(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Panicln(err)
		//w.WriteHeader(http.StatusInternalServerError)
		//return
	}

	//客户端发送的请求的内容长度大于0 并且当前是选举出来了节点即进行分布式集群同步
	if len(r.Form["message"]) > 0 && rf.currentLeader != "-1" {
		msg := r.Form["message"][0]
		m := new(Message)
		m.MsgID = getRandom()
		m.MsgBody = msg

		//http端口监听到消息后转发给leader处理;
		//fmt.Println("http监听到了消息，准备发送给领导者，消息id:", m.msgID)
		port := nodePool[rf.currentLeader]
		rp, err := rpc.DialHTTP("tcp", "127.0.0.1"+port)
		if err != nil {
			log.Panicln(err)
		}

		var b bool
		err = rp.Call("Raft.LeaderReceiveMessage", m, &b)
		if err != nil {
			log.Panicln(err)
		}
		fmt.Println("消息是否已经发送到领导者:", b)
		//w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok!!!"))
	}
}

//返回一个十位数的随机数，作为msg.id
func getRandom() int {
	id := rand.Intn(1000000000) + 1000000000
	return id
}
