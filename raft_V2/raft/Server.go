package raft

import (
	"fmt"
	"log"
	"os"
)

//节点总数; 默认是三个
var nodeCount = 3

//所有的节点列表
var nodePool map[string]string

//线程锁
//var mutex sync.Mutex

//选举超时时间
var electionTimeout = 3

//心跳检测超时时间
var heartbeatTimeout = 7

//心跳检测频率; 每隔多少秒进行检测一次
var heartbeatRate = 3

//存储客户端发送来的消息
var messageStore = make(map[int]string)

func Start(nodeSum int, nodeTable map[string]string) {
	nodePool = nodeTable
	nodeCount = nodeSum
	if len(os.Args) < 2 {
		log.Panicln("程序参数不正确")
	}

	id := os.Args[1]

	fmt.Println(id, nodeTable[id])
	//创建raft实例
	raft := NewRaft(id, nodeTable[id])

	//注册rpc服务; 节点的状态转换需要RPC通信; 包括选举、日志的复制等
	go rpcRegister(raft)

	//开启向其他节点发送心跳信息;
	go raft.startHeartbeat()

	//开启一个http监听客户端发来的信息
	if id == "A" {
		fmt.Println(id)
		go raft.httpListen()
	}

	//开启选举;
	//初始时各个节点都是follower节点,所以每个节点都同时发起选举变为candidate状态();
	go raft.tryTobeCandidateAndElection()

	//所有的follower节点都进行心跳信息超时检测; 看leader节点是否down掉
	go raft.detectLeaderDown()

	select {}
}
