package cmd_handler

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"hero_story.go_server/biz_server/base"
	"hero_story.go_server/biz_server/mod/login/login_srv"
	"hero_story.go_server/biz_server/mod/user/user_data"
	"hero_story.go_server/biz_server/msg"
	"hero_story.go_server/comm/log"
)

func init() {
	cmdHandlerMap[uint16(msg.MsgCode_USER_LOGIN_CMD.Number())] = handleUserLoginCmd
}

// 用户登录指令处理器
func handleUserLoginCmd(ctx base.MyCmdContext, pbMsgObj *dynamicpb.Message) {
	if nil == ctx ||
		nil == pbMsgObj {
		return
	}

	userLoginCmd := &msg.UserLoginCmd{}

	pbMsgObj.Range(func(f protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		userLoginCmd.ProtoReflect().Set(f, v)
		return true
	})

	log.Info(
		"收到用户登录消息! userName = %s, password = %s",
		userLoginCmd.GetUserName(),
		userLoginCmd.GetPassword(),
	)

	// 根据用户名称和密码登录
	bizResult := login_srv.LoginByPasswordAsync(userLoginCmd.GetUserName(), userLoginCmd.GetPassword())

	if nil == bizResult {
		log.Error(
			"业务结果返回空值, userName = %s",
			userLoginCmd.GetUserName(),
		)
		return
	}

	// 执行了一大堆别的操作...

	bizResult.OnComplete(func() {
		returnedObj := bizResult.GetReturnedObj()

		if nil == returnedObj {
			log.Error(
				"用户不存在, userName = %s",
				userLoginCmd.GetUserName(),
			)
			return
		}

		user := returnedObj.(*user_data.User)

		userLoginResult := &msg.UserLoginResult{
			UserId:     uint32(user.UserId),
			UserName:   user.UserName,
			HeroAvatar: user.HeroAvatar,
		}

		ctx.BindUserId(user.UserId)
		ctx.Write(userLoginResult)
	})
}
