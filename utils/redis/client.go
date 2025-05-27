package rediswqj

import (
	"fmt"
	"github.com/go-redis/redis"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func RedisConn(keyName, hostPort, redisPass string) string {
	//hostPort = "172.16.62.128:30082"
	log.Log.Info("redis connect")
	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{hostPort},
		DB:            15,
		Password:      redisPass,
	})
	res := rdb.Get(keyName)
	log.Log.Info("redis key " + keyName + ": " + res.Val())

	err := rdb.Close()
	if err != nil {
		return ""
	}
	return res.Val()
}

func SlaveIsReady(ip, port, password string) (string, error) {
	options := &redis.Options{
		Addr:     net.JoinHostPort(ip, port),
		Password: password,
		DB:       0,
	}
	rClient := redis.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info("replication").Result()
	if err != nil {
		return info, err
	}
	return info, nil
}

func FailOverOperate(ip, mastername string) (string, error) {
	options := &redis.Options{
		Addr:     net.JoinHostPort(ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := redis.NewClient(options)
	defer rClient.Close()

	cmd := redis.NewStringCmd("SENTINEL", "failover", mastername)
	err := rClient.Process(cmd)
	if err != nil {
		return "failover failed", err
	}

	res, err := cmd.Result()
	if err != nil {
		return "failover failed", err
	}
	return fmt.Sprintf("failover succeeded: %s", res), nil
}
