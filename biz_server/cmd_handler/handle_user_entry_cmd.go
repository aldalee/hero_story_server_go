package cmd_handler

import (
	"context"
	"fmt"
	"strconv"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/types/dynamicpb"
	"hero_story.go_server/biz_server/base"
	"hero_story.go_server/biz_server/mod/user/user_dao"
	"hero_story.go_server/biz_server/mod/user/user_data"
	"hero_story.go_server/biz_server/msg"
	"hero_story.go_server/biz_server/network/broadcaster"
	"hero_story.go_server/comm/async_op"
	"hero_story.go_server/comm/cluster"
	"hero_story.go_server/comm/log"
)

func init() {
	cmdHandlerMap[uint16(msg.MsgCode_USER_ENTRY_CMD.Number())] = handleUserEntryCmd
}

// 用户入场指令处理器
func handleUserEntryCmd(ctx base.MyCmdContext, _ *dynamicpb.Message) {
	if nil == ctx ||
		ctx.GetUserId() <= 0 {
		return
	}

	log.Info(
		"收到用户入场消息! userId = %d",
		ctx.GetUserId(),
	)

	async_op.Process(
		int(ctx.GetUserId()),

		//
		// 执行异步过程
		func() {
			user := user_dao.GetUserById(ctx.GetUserId())

			if nil == user {
				log.Error("数据库中不存在该用户! userId = %d", ctx.GetUserId())
				return
			}

			user_data.GetUserGroup().Add(user)

			etcdKey := fmt.Sprintf(
				"user_%d_at_game_server_id",
				ctx.GetUserId(),
			)

			etcdVal := strconv.Itoa(base.G_serverId)

			cluster.GetEtcdCli().Put( // redis.set
				context.TODO(),
				etcdKey,
				etcdVal,
				clientv3.WithLease(base.G_leaseIdForServerLife), // redis.expire TTL
			)

			// Etcd 设置过期时间是通过租约 Id ( LeaseId )
		},

		//
		// 回到主线程之后执行
		func() {
			// 获取用户数据
			user := user_data.GetUserGroup().GetByUserId(ctx.GetUserId())

			if nil == user {
				log.Error(
					"未找到用户数据, userId = %d",
					ctx.GetUserId(),
				)
				return
			}

			userEntryResult := &msg.UserEntryResult{
				UserId:     uint32(ctx.GetUserId()),
				UserName:   user.UserName,
				HeroAvatar: user.HeroAvatar,
			}

			broadcaster.Broadcast(userEntryResult)
		},
	)
}
