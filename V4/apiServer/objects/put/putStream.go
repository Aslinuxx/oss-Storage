package put

import (
	"Storage/V4/apiServer/heartbeat"
	"Storage/lib/objectstream"
	"fmt"
)

func putStream(hash string, size int64) (*objectstream.TempPutStream, error) {
	server := heartbeat.ChooseRandomDataServer()

	if server == "" {
		return nil, fmt.Errorf("cannot find any dataServer")
	}

	return objectstream.NewTempPutStream(server, hash, size)
}
