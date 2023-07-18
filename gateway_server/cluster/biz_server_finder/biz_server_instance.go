package biz_server_finder

import (
	"github.com/gorilla/websocket"
	"hero_story.go_server/biz_server/base"
	"sync"
)

// key = ServerJobType, val = Map<ServerId, BizServerInstance>
var bizServerInstanceMap = &sync.Map{}

type BizServerInstance struct {
	*base.BizServerData
	*websocket.Conn
}

func addBizServerInstance(bizServerInstance *BizServerInstance) {
	if nil == bizServerInstance ||
		nil == bizServerInstance.SjtArray ||
		len(bizServerInstance.SjtArray) <= 0 {
		return
	}

	for _, sjt := range bizServerInstance.SjtArray {
		innerMap, ok := bizServerInstanceMap.Load(sjt)

		if !ok {
			innerMap = &sync.Map{}
			bizServerInstanceMap.LoadOrStore(sjt, innerMap)
		}

		innerMap, ok = bizServerInstanceMap.Load(sjt)

		if !ok {
			panic("内置字典为空")
		}

		innerMap.(*sync.Map).Store(
			bizServerInstance.ServerId,
			bizServerInstance,
		)
	}
}

func deleteBizServerInstance(bizServerInstance *BizServerInstance) {
	if nil == bizServerInstance ||
		nil == bizServerInstance.SjtArray ||
		len(bizServerInstance.SjtArray) <= 0 {
		return
	}

	for _, sjt := range bizServerInstance.SjtArray {
		innerMap, ok := bizServerInstanceMap.Load(sjt)

		if !ok {
			continue
		}

		innerMap.(*sync.Map).Delete(bizServerInstance.ServerId)
	}
}
