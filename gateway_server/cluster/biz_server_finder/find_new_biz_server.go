package biz_server_finder

import (
	"context"
	"encoding/json"

	clientv3 "go.etcd.io/etcd/client/v3"
	"hero_story.go_server/biz_server/base"
	"hero_story.go_server/comm/cluster"
)

// FindNewBizServer 查找新的业务服务器,
// 通过 Etcd 的 watch 指令监听 "hero_story.go_server/biz_server" 为前缀的所有关键字的变化
func FindNewBizServer() {
	etcdCli := cluster.GetEtcdCli()

	go func() {
		watchChan := etcdCli.Watch(context.TODO(), "hero_story.go_server/biz_server", clientv3.WithPrefix())

		for resp := range watchChan {
			for _, event := range resp.Events {
				switch event.Type {
				case 0: // PUT
					//log.Info("发现新的业务服务器, " + string(event.Kv.Value))

					serverData := &base.BizServerData{}
					_ = json.Unmarshal(event.Kv.Value, serverData)

					connToBizServer(serverData)
				case 1: // DELETE
				}
			}
		}
	}()
}
