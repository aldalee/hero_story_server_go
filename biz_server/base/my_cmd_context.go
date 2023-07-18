package base

import "google.golang.org/protobuf/reflect/protoreflect"

// MyCmdContext 自定义得指令处理器上下文
// 类似于 Java Netty 中的 ChannelHandlerContext
type MyCmdContext interface {
	// BindUserId 绑定用户 Id
	BindUserId(val int64)

	// GetUserId 获取用户 Id
	GetUserId() int64

	// GetClientIpAddr 获取客户端 IP 地址
	GetClientIpAddr() string

	// Write 写出消息对象
	Write(msgObj protoreflect.ProtoMessage)

	// SendError 发送错误消息
	SendError(errorCode int, errorInfo string)

	// Disconnect 断开连接
	Disconnect()
}
