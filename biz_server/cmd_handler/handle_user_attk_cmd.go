package cmd_handler

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"hero_story.go_server/biz_server/base"
	"hero_story.go_server/biz_server/mod/user/user_data"
	"hero_story.go_server/biz_server/mod/user/user_lso"
	"hero_story.go_server/biz_server/msg"
	"hero_story.go_server/biz_server/network/broadcaster"
	"hero_story.go_server/comm/lazy_save"
)

func init() {
	cmdHandlerMap[uint16(msg.MsgCode_USER_ATTK_CMD.Number())] = handleUserAttkCmd
}

// 用户攻击指令处理器
func handleUserAttkCmd(ctx base.MyCmdContext, pbMsgObj *dynamicpb.Message) {
	if nil == ctx ||
		nil == pbMsgObj {
		return
	}

	userAttkCmd := &msg.UserAttkCmd{}

	pbMsgObj.Range(func(f protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		userAttkCmd.ProtoReflect().Set(f, v)
		return true
	})

	userAttkResult := &msg.UserAttkResult{
		AttkUserId:   uint32(ctx.GetUserId()),
		TargetUserId: userAttkCmd.TargetUserId,
	}

	broadcaster.Broadcast(userAttkResult)

	user := user_data.GetUserGroup().GetByUserId(int64(userAttkCmd.GetTargetUserId()))

	if nil == user {
		return
	}

	var subtractHp int32 = 10
	user.CurrHp -= subtractHp

	userSubtractHpResult := &msg.UserSubtractHpResult{
		SubtractHp:   uint32(subtractHp),
		TargetUserId: userAttkCmd.TargetUserId,
	}

	broadcaster.Broadcast(userSubtractHpResult)

	lso := user_lso.GetUserLso(user)

	// 执行延迟保存
	lazy_save.SaveOrUpdate(lso)
}
