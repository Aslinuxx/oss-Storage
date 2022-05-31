package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"time"
)

//实现rpc服务

func rpcRegister(raft *Raft) {
	//注册一个rpc服务器
	//err := rpc.Register(raft)
	//if err != nil {
	//	log.Panicln("注册RPC服务失败了", err)
	//}
	//
	//port := raft.node.Port
	////fmt.Println(port)
	////把rpc服务绑定到http协议上
	//rpc.HandleHTTP()
	////监听端口
	//err = http.ListenAndServe(":"+port, nil)
	//if err != nil {
	//	fmt.Println("注册RPC服务失败", err)
	//}
	//log.Println("注册RPC服务成功")

	//注册一个RPC服务器
	if err := rpc.Register(raft); err != nil {
		log.Panicln("注册RPC失败", err)
	}
	port := raft.node.Port
	//把RPC服务绑定到http协议上
	rpc.HandleHTTP()
	//127.0.0.1:6870|6871|6872
	http.ListenAndServe(port, nil)

	log.Println("RPC注册成功!")

}

//通过RPC服务向其他节点进行广播
func (rf *Raft) broadCast(method string, args interface{}, fun func(ok bool)) {
	////不给自己广播
	//for nodeId, nodePort := range nodePool {
	//	if nodeId == rf.me {
	//		continue
	//	}
	//
	//	//连接远程节点的rpc
	//	rp, err := rpc.DialHTTP("tcp", "127.0.0.1"+nodePort)
	//	if err != nil {
	//		fun(false)
	//		continue
	//	}
	//	var bo = false
	//	err = rp.Call(method, args, &bo) //调用method方法、传入参数、返回值(可以是一个反射)
	//	if err != nil {
	//		fun(false)
	//		continue
	//	}
	//	fun(bo)
	//
	//}

	for nodeID, nodePort := range nodePool {
		//不广播自己
		if nodeID == rf.me {
			continue
		}
		//连接远程节点的rpc
		conn, err := rpc.DialHTTP("tcp", "127.0.0.1"+nodePort)
		if err != nil {
			//连接失败，调用回调
			fun(false)
			continue
		}
		var bo bool
		err = conn.Call(method, args, &bo)
		if err != nil {
			//调用失败，调用回调
			fun(false)
			continue
		}
		//回调
		fun(bo)
	}
}

//当该节点收到其他节点的心跳信息的回复信息
//除了选举的时候,后序节点收到的心跳信息都是leader节点发来的
func (rf *Raft) heartbeatResponse(node NodeInfo, b *bool) error {
	//因为发送心跳的一定是leader，之所以写这一句的目的是如果有down的节点恢复了，直接是follower，所以直接告诉它leader是谁即可
	rf.setCurrentLeader(node.ID)
	rf.lastReceiveHeartbeatTime = millisecond()
	fmt.Printf("收到来自leader[%s]节点的心跳检测\n", node.ID)

	*b = true
	return nil
}

//leader节点接收并转发日志至follower节点
func (rf *Raft) LeaderReceiveMessage(msg Message, b *bool) error {
	fmt.Println("leader节点接收到转发过来的信息,id为:%d\n", msg.msgID)
	messageStore[msg.msgID] = msg.msgBody // leader节点自己存储该条信息
	*b = true
	fmt.Println("准备将消息进行广播...")
	//广播给其他跟随者
	num := 0
	go rf.broadCast("Raft.ReceiveMessage", msg, func(ok bool) {
		if ok {
			num++
		}
	})

	for {
		//如果超过半数收到消息,则说明日志已经正确发送到follower节点
		if num > nodeCount/2-1 {
			fmt.Printf("全网已超过半数节点接收到消息id：%d\nraft验证通过,可以打印消息,id为：%d\n", msg.msgBody, msg.msgID)
			fmt.Println("消息为：", messageStore[msg.msgID], "\n")
			rf.lastSendMessageTime = millisecond()
			fmt.Println("准备将消息提交信息发送至客户端...")
			go rf.broadCast("Raft.ConfirmedMessage", msg, func(ok bool) {
			})
			break
		} else {
			//可能别的节点还没回复，等待一会
			time.Sleep(time.Millisecond * 100)
		}
	}
	return nil
}

//follower节点接收leader节点同步的消息,并将该消息存储下来
func (rf *Raft) ReceiveMessage(msg Message, b *bool) error {
	fmt.Println("接收到leader同步的日志,消息id为:%d\n", msg.msgID)
	messageStore[msg.msgID] = msg.msgBody
	*b = true
	fmt.Println("已成功同步leader发来的日志,待leader确认后打印")
	return nil
}

//日志已经同步到半数以上的follower节点上,leader节点准备回复客户端已收到该信息; 并打印
func (rf *Raft) ConfirmedMessage(msg Message, b *bool) error {
	go func() {
		for {
			if _, ok := messageStore[msg.msgID]; ok {
				fmt.Printf("raft验证通过,可以打印消息,id为:[%d],消息为:[%s]\n", msg.msgID, messageStore[msg.msgID])
				rf.lastSendMessageTime = millisecond()
				break
			} else {
				//可能这个节点的网络传输很慢，等一会
				time.Sleep(time.Millisecond * 10)
			}
		}
	}()
	*b = true
	return nil
}

//节点向candidate节点进行投票;
//投票的策略是各个节点只能投一票;已投过不能再投
func (rf *Raft) Vote(node NodeInfo, b *bool) error {
	//fmt.Println("开始投票")
	//if rf.voteFor == "-1" && rf.currentLeader == "-1" {
	//	rf.setVoteFor(node.ID)
	//	fmt.Println("投票成功,已投%s节点\n", node.ID)
	//	*b = true
	//} else {
	//	*b = false
	//}
	//return nil

	if rf.voteFor == "-1" && rf.currentLeader == "-1" {
		rf.setVoteFor(node.ID)
		fmt.Printf("投票成功，已投%s节点\n", node.ID)
		*b = true
	} else {
		*b = false
	}
	return nil
}

//产生新的leader后,向follower进行广播新的leader
func (rf *Raft) ConfirmedLeader(id string, b *bool) error {
	rf.setCurrentLeader(id)
	*b = true
	fmt.Println("已产生新的leader,ID为:", id)
	rf.reDefault()
	return nil
}
