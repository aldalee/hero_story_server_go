package cluster

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var etcdCli *clientv3.Client

func InitEtcdCli(etcdEndpointArray []string) {
	if nil == etcdEndpointArray ||
		len(etcdEndpointArray) <= 0 {
		return
	}

	var err error

	etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpointArray,
		DialTimeout: 5 * time.Second,
	})

	if nil != err {
		panic(err)
	}
}

func GetEtcdCli() *clientv3.Client {
	return etcdCli
}
