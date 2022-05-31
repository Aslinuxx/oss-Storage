package heartbeat

import (
	"math/rand"
)

func ChooseRandomDataServer() string {
	ds := getDataServers()

	n := len(ds)
	if n == 0 {
		return ""
	}

	return ds[rand.Intn(n)]
}

func getDataServers() []string {
	mutex.Lock()
	ds := make([]string, 0)
	for s, _ := range dataServers {
		ds = append(ds, s)
	}

	mutex.Unlock()

	return ds

}