package websocket

import (
	"context"
	"fmt"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"hero_story.go_server/comm/cluster"
	"hero_story.go_server/comm/log"
)

func (cmdCtx *CmdContextImpl) BindUserId(val int64) {
	etcdCli := cluster.GetEtcdCli()

	// 把用户 Id 和网关服 Id 写入到 Etcd 里,
	// 写入是有条件的!
	// 1、 必须通过事务来写!
	// 2、 写的时候, 必须不存在指定的关键字!
	// 避免出现两个客户端同时连接的极端情况

	etcdKey := fmt.Sprintf("user_%d_at_gateway_server_id", val)
	etcdVal := strconv.Itoa(int(cmdCtx.GatewayServerId))

	// 生成一个 10 秒钟自动过期的租约
	grantResp, err := etcdCli.Grant(context.TODO(), 10)

	if nil != err {
		log.Error("%+v", err)
		return
	}

	txnResp, err := etcdCli.Txn(context.TODO()).If(
		clientv3.Compare(clientv3.CreateRevision(etcdKey), "=", 0),
	).Then(
		clientv3.OpPut(
			etcdKey, etcdVal,
			clientv3.WithLease(grantResp.ID),
		),
	).Commit()

	if nil != err {
		log.Error("%+v", err)
		return
	}

	if !txnResp.Succeeded {
		log.Error("未能绑定用户 Id, Etcd 事务提交失败!, userId = %d", val)
		return
	}

	go func() {
		for {
			if cmdCtx.closeFlag {
				break
			}

			time.Sleep(5 * time.Second)
			_, _ = etcdCli.KeepAliveOnce(context.TODO(), grantResp.ID)
		}
	}()

	cmdCtx.userId = val
}
