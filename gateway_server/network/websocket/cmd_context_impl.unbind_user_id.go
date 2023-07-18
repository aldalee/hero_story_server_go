package websocket

import (
	"context"
	"fmt"
	"strconv"

	clientv3 "go.etcd.io/etcd/client/v3"
	"hero_story.go_server/comm/cluster"
	"hero_story.go_server/comm/log"
)

// 解绑用户 Id
func (ctx *CmdContextImpl) UnbindUserId() {
	UnbindUserId(ctx.userId, ctx.GatewayServerId)
	ctx.userId = 0
}

// 根据用户 Id 和网关服 Id 执行解绑逻辑
func UnbindUserId(userId int64, gatewayServerId int32) {
	if userId <= 0 ||
		gatewayServerId <= 0 {
		return
	}

	etcdKey := fmt.Sprintf("user_%d_at_gateway_server_id", userId)
	etcdVal := strconv.Itoa(int(gatewayServerId))

	etcdCli := cluster.GetEtcdCli()
	txnResp, err := etcdCli.Txn(context.TODO()).If(
		clientv3.Compare(clientv3.Value(etcdKey), "=", etcdVal),
	).Then(
		clientv3.OpDelete(
			etcdKey,
		),
	).Commit()

	if nil != err {
		log.Error("%+v", err)
		return
	}

	if txnResp.Succeeded {
		log.Info("事务执行成功, 解绑用户 Id, userId = %d", userId)
	} else {
		log.Info("事务执行失败, 未解绑用户 Id")
	}
}
