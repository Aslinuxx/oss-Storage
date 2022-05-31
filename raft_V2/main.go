package main

import "raft_V2/raft"

func main() {
	//分布式集群节点列表
	nodeTable := map[string]string{
		"A": ":6870",
		"B": ":6871",
		"C": ":6872",
	}

	raft.Start(3, nodeTable)
}
