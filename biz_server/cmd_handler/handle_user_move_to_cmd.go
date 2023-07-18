package cmd_handler

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"hero_story.go_server/biz_server/base"
	"hero_story.go_server/biz_server/mod/user/user_data"
	"hero_story.go_server/biz_server/msg"
	"hero_story.go_server/biz_server/network/broadcaster"
	"time"
)

func init() {
	cmdHandlerMap[uint16(msg.MsgCode_USER_MOVE_TO_CMD.Number())] = handleUserMoveToCmd
}

// 用户移动到指令处理器
func handleUserMoveToCmd(ctx base.MyCmdContext, pbMsgObj *dynamicpb.Message) {
	if nil == ctx ||
		ctx.GetUserId() <= 0 ||
		nil == pbMsgObj {
		return
	}

	// 获取用户数据
	user := user_data.GetUserGroup().GetByUserId(ctx.GetUserId())

	if nil == user {
		return
	}

	userMoveToCmd := &msg.UserMoveToCmd{}

	pbMsgObj.Range(func(f protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		userMoveToCmd.ProtoReflect().Set(f, v)
		return true
	})

	if nil == user.MoveState {
		user.MoveState = &user_data.MoveState{}
	}

	nowTime := time.Now().UnixMilli()

	user.MoveState.FromPosX = userMoveToCmd.MoveFromPosX
	user.MoveState.FromPosY = userMoveToCmd.MoveFromPosY
	user.MoveState.ToPosX = userMoveToCmd.MoveToPosX
	user.MoveState.ToPosY = userMoveToCmd.MoveToPosY
	user.MoveState.StartTime = nowTime

	userMoveToResult := &msg.UserMoveToResult{
		MoveUserId:    uint32(ctx.GetUserId()),
		MoveFromPosX:  userMoveToCmd.MoveFromPosX,
		MoveFromPosY:  userMoveToCmd.MoveFromPosY,
		MoveToPosX:    userMoveToCmd.MoveToPosX,
		MoveToPosY:    userMoveToCmd.MoveToPosY,
		MoveStartTime: uint64(nowTime),
	}

	broadcaster.Broadcast(userMoveToResult)
}
