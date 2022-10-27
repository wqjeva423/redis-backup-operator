package rediswqj

import (
	"github.com/go-redis/redis"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func RedisConn(keyName, hostPort string) string {
	//hostPort = "172.16.59.125:30082"
	log.Log.Info("redis connect")
	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{hostPort},
		DB:            15,
		Password:      "root",
	})
	res := rdb.Get(keyName)
	log.Log.Info("redis key " + keyName + ": " + res.Val())

	err := rdb.Close()
	if err != nil {
		return ""
	}
	return res.Val()
}
