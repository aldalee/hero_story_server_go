package websocket

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"hero_story.go_server/biz_server/msg"
	"hero_story.go_server/comm/log"
)

type innerCmdContextImpl struct {
	gatewayServerId int32
	remoteSessionId int32
	userId          int64
	clientIpAddr    string
	lastActiveTime  int64
	*GatewayServerConn
}

// BindUserId 绑定用户 Id
func (ctx *innerCmdContextImpl) BindUserId(val int64) {
	ctx.userId = val
}

// GetUserId 获取用户 Id
func (ctx *innerCmdContextImpl) GetUserId() int64 {
	return ctx.userId
}

// GetClientIpAddr 获取客户端 IP 地址
func (ctx *innerCmdContextImpl) GetClientIpAddr() string {
	return ctx.clientIpAddr
}

// Write 写出消息对象
func (ctx *innerCmdContextImpl) Write(msgObj protoreflect.ProtoMessage) {
	if nil == msgObj {
		return
	}

	byteArray, err := msg.Encode(msgObj)

	if nil != err {
		log.Error("%+v", err)
		return
	}

	innerMsg := &msg.InternalServerMsg{
		GatewayServerId: ctx.gatewayServerId,
		SessionId:       ctx.remoteSessionId,
		UserId:          ctx.userId,
		MsgData:         byteArray,
	}

	ctx.GatewayServerConn.sendMsgQ <- innerMsg
}

// SendError 发送错误消息
func (ctx *innerCmdContextImpl) SendError(errorCode int, errorInfo string) {
}

// Disconnect 断开连接
func (ctx *innerCmdContextImpl) Disconnect() {
	innerMsg := &msg.InternalServerMsg{
		GatewayServerId: ctx.gatewayServerId,
		SessionId:       ctx.remoteSessionId,
		UserId:          ctx.userId,
		Disconnect:      1,
	}

	ctx.GatewayServerConn.sendMsgQ <- innerMsg
}
