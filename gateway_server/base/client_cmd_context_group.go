package base

import (
	"sync"
)

// ClientCmdContextGroup 客户端指令上下文分组
type ClientCmdContextGroup struct {
	innerMap *sync.Map // ConcurrentHashMap<int, CmdContextImpl>
}

var cmdContextImplGroupInstance = &ClientCmdContextGroup{
	innerMap: &sync.Map{},
}

// GetCmdContextImplGroup 获取单例对象
func GetCmdContextImplGroup() *ClientCmdContextGroup {
	return cmdContextImplGroupInstance
}

// Add 添加客户端指令上下文
func (group *ClientCmdContextGroup) Add(sessionId int32, cmdCtx ClientCmdContext) {
	if nil == cmdCtx {
		return
	}

	group.innerMap.Store(sessionId, cmdCtx)
}

// RemoveBySessionId 根据会话 Id 移除客户端指令上下文
func (group *ClientCmdContextGroup) RemoveBySessionId(sessionId int32) {
	if sessionId <= 0 {
		return
	}

	group.innerMap.Delete(sessionId)
}

// GetBySessionId 根据会话 Id 获取客户端指令上下文
func (group *ClientCmdContextGroup) GetBySessionId(sessionId int32) ClientCmdContext {
	if sessionId <= 0 {
		return nil
	}

	cmdCtx, ok := group.innerMap.Load(sessionId)

	if !ok {
		return nil
	}

	return cmdCtx.(ClientCmdContext)
}

func (group *ClientCmdContextGroup) GetByUserId(userId int64) ClientCmdContext {
	if userId <= 0 {
		return nil
	}

	var findCmdCtx ClientCmdContext

	group.innerMap.Range(func(_, val interface{}) bool {
		if val.(ClientCmdContext).GetUserId() == userId {
			findCmdCtx = val.(ClientCmdContext)
			return false
		}

		return true
	})

	return findCmdCtx
	// 改完之后别忘了回去修改 gateway_server.go 代码；
}
