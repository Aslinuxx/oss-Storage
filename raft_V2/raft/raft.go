package raft

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	follower  = 0
	candidate = 1
	leader    = 2
)

//每个集群节点的信息;节点的编号、地址
type NodeInfo struct {
	ID   string
	Port string
}

//raft节点的具体数据结构
type Raft struct {
	node                     *NodeInfo
	vote                     int        // 本节点获得的投票数
	mutex                    sync.Mutex // 线程锁
	me                       string     //节点编号
	currentTerm              int        //当前任期
	voteFor                  string     //投票给哪个节点
	status                   int        //当前节点的状态;节点有三个状态:1.follower(跟随者) 2.candidate(候选者) 3.leader(领导者)
	lastSendMessageTime      int64      //发送最后一条消息的时间
	lastSendHeartbeatTime    int64      //发送最后一次心跳信息的时间
	lastReceiveHeartbeatTime int64      // 最近一次收到的心跳信息的时间; 如果选出主节点后,该时间就是follower节点收到leader节点心跳信息的时间
	currentLeader            string     //当前的领导者节点
	heartbeatTimeout         int        //心跳超时时间
	voteChan                 chan bool  //投票channel
	heartbeatChan            chan bool  //心跳信息
}

//创建一个raft实例;并进行初始化
func NewRaft(id, port string) *Raft {
	rf := new(Raft)
	//初始化
	//初始化的时候,每个节点都是follower,都没有获得投票;初始任期Term都是0
	rf.node = &NodeInfo{id, port}

	rf.me = id
	rf.setVote(0)                          //自己节点获得的投票数
	rf.setVoteFor("-1")                    //初始化的时候,不给任何节点投票(包括自己)
	rf.setStatus(follower)                 // 每一节点的初始状态都是follower 跟随者
	rf.lastSendHeartbeatTime = 0           //最后发送一次心跳检测时间
	rf.lastReceiveHeartbeatTime = 0        // 最后一次收到leader节点心跳信息的时间
	rf.heartbeatTimeout = heartbeatTimeout //心跳超时时间
	rf.setCurrentLeader("-1")              //初始的时候没有领导者
	rf.setTerm(0)                          // 初始的任期都是0
	rf.voteChan = make(chan bool)          //channel需要初始化才能使用
	rf.heartbeatChan = make(chan bool)

	log.Println("创建raft实例成功")

	return rf
}

//为当前节点设置已获得的投票数
func (rf *Raft) setVote(num int) {
	rf.mutex.Lock()
	rf.vote = num
	rf.mutex.Unlock()
}

//向某一节点投票
func (rf *Raft) setVoteFor(id string) {
	rf.mutex.Lock()
	rf.voteFor = id
	rf.mutex.Unlock()
}

//设置当前节点的状态
func (rf *Raft) setStatus(status int) {
	rf.mutex.Lock()
	rf.status = status
	rf.mutex.Unlock()
}

func (rf *Raft) setCurrentLeader(id string) {
	rf.mutex.Lock()
	rf.currentLeader = id
	rf.mutex.Unlock()
}

func (rf *Raft) setTerm(term int) {
	rf.mutex.Lock()
	rf.currentTerm = term
	rf.mutex.Unlock()

}

//自己的获票加1
func (rf *Raft) voteAdd() {
	rf.mutex.Lock()
	rf.vote++
	rf.mutex.Unlock()
}

//自己的任期加1
func (rf *Raft) termAdd() {
	rf.mutex.Lock()
	rf.currentTerm++
	rf.mutex.Unlock()
}

func (rf *Raft) reDefault() {
	//mutex.Lock()
	rf.setStatus(follower)
	rf.setVote(0)
	rf.setVoteFor("-1")
	//mutex.Unlock()

}

//产生随机值
func randRange(min, max int64) int64 {
	//用于心跳信号的时间
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}

//获取当前时间;以毫秒数表示
func millisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//向其他follower节点发送心跳信息
func (rf *Raft) startHeartbeat() {
	<-rf.heartbeatChan //如果通道开启就会向其他节点进行发送心跳信息;如果chan没有打开,则会阻塞等待
	for {              //这里是无限循环一直广播发送心跳信息到follower节点
		fmt.Println("本节点开始发送心跳信息")
		rf.broadCast("Raft.HeartbeatResponse", rf.node, func(ok bool) {
			fmt.Println("收到回复", ok)
		})
		//最后一次心跳的时间
		rf.lastSendHeartbeatTime = millisecond()
		time.Sleep(time.Second * time.Duration(heartbeatRate))

	}

}

//follower节点尝试变为candidate节点,并进行选举
func (rf *Raft) tryTobeCandidateAndElection() {
	for {
		//如果成功变为candidate节点后就进行尝试向其他节点要选票进行选举
		if rf.becomeCandidate() {
			if rf.election() {
				break
			} else {
				continue //选举超时,重新进行一次选举;说明该次选举不成功,没有产生leader
			}
		} else {
			//当前节点没有变为候选人,则退出竞选leader
			break
		}
	}
}

func (rf *Raft) becomeCandidate() bool {
	//等待一定随机时间后,开始变为candidate节点;这样是防止所有的follower节点同时发起选举
	r := randRange(1500, 5000)
	time.Sleep(time.Duration(r) * time.Millisecond)

	//如果当前节点已经投过票,或者已经存在领导者,或者当前节点不是follower节点,则不用变为candidate节点
	if rf.status == follower && rf.currentLeader == "-1" && rf.voteFor == "-1" {
		rf.setStatus(candidate) //当前节点变为candidate
		rf.setVoteFor(rf.me)    //投票给自己
		rf.voteAdd()            //自己的获票数量加1
		rf.termAdd()            //任期加1
		fmt.Println("本节点已经变成候选人状态")
		fmt.Printf("当前获得的票数：%d\n", rf.vote)

		//开启选举通道;即变为candidate节点后向其他节点发起选举请求;

		return true
	}
	return false
}

//candidate节点向其他节点要票,进行选举变为leader节点
func (rf *Raft) election() bool {
	fmt.Println("开始进行leader竞选,向其他节点进行广播")
	go rf.broadCast("Raft.Vote", rf.node, func(ok bool) {
		rf.voteChan <- ok // voteChan接收其他节点的投票;
		fmt.Println(ok)
	})

	//统计票数
	for {
		select {
		case <-time.After(time.Second * time.Duration(electionTimeout)):
			fmt.Println("领导者选举超时，节点变更为追随者状态")
			rf.reDefault()
			return false
		case ok := <-rf.voteChan:
			if ok {
				rf.voteAdd()
				fmt.Printf("获得来自其他节点的投票，当前得票数：%d\n", rf.vote)
			}
			if rf.vote >= nodeCount/2+1 && rf.currentLeader == "-1" {
				fmt.Println("获得大多数节点的同意，本节点被选举成为了leader")
				//节点状态变为2，代表leader
				rf.setStatus(leader)
				//当前领导者为自己
				rf.setCurrentLeader(rf.me)
				fmt.Println("向其他节点进行广播本节点成为了leader...")
				go rf.broadCast("Raft.ConfirmedLeader", rf.me, func(ok bool) {
					fmt.Println("其他节点:是否同意[", rf.me, "]为领导者", ok)
				})
				//有leader了，可以发送心跳包了
				rf.heartbeatChan <- true // leader的heartbeatChan是一个无缓冲区的channel;所以这里leader在chan中放入true在 StartHeartbeat()中被读出;所以使该管道被激活
				return true
			}
		}

	}
}

//follower节点检测leader节点的是否down掉,根据最后接收到的leader节点的心跳信息时间来看
func (rf *Raft) detectLeaderDown() {
	for {
		//1秒检测一次
		time.Sleep(time.Second)
		if rf.status == follower && rf.currentLeader != "-1" && (millisecond()-rf.lastReceiveHeartbeatTime) > int64(rf.heartbeatTimeout*1000) {
			fmt.Printf("心跳检测超时,已超过%d秒\n", rf.heartbeatTimeout)
			fmt.Println("即将开始重新选举")
			rf.reDefault()
			rf.setCurrentLeader("-1")
			rf.lastReceiveHeartbeatTime = 0
			go rf.tryTobeCandidateAndElection()
		}
	}

}
