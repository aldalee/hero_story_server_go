package cmd_handler

import (
	"google.golang.org/protobuf/types/dynamicpb"
	"hero_story.go_server/biz_server/base"
)

// CmdHandlerFunc 自定义的消息处理函数
type CmdHandlerFunc func(ctx base.MyCmdContext, pbMsgObj *dynamicpb.Message)

// 指令处理器字典, key = msgCode, val = CmdHandlerFunc
var cmdHandlerMap = make(map[uint16]CmdHandlerFunc)

// CreateCmdHandler 根据消息代号创建指令处理器
func CreateCmdHandler(msgCode uint16) CmdHandlerFunc {
	return cmdHandlerMap[msgCode]
}
