package base

//
// ClientCmdContext 客户端指令上下文
type ClientCmdContext interface {
	//
	// BindUserId 绑定用户 Id
	BindUserId(val int64)

	//
	// GetUserId 获取用户 Id
	GetUserId() int64

	//
	// GetClientIpAddr 获取客户端 IP 地址
	GetClientIpAddr() string

	//
	// Write 写出消息
	Write(byteArray []byte)

	//
	// SendError 发送错误
	SendError(errorCode int, errorInfo string)

	//
	// Disconnect 断开连接
	Disconnect()
}
