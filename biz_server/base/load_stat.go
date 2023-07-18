package base

import (
	"sync"
	"sync/atomic"
)

type LoadStat struct {
	// 用户 Id 字典,
	// key = gatewayServerId, val = Map<userId, bool>
	userIdMap *sync.Map

	// 总人数
	totalCount int32
}

// 单例对象
var loadStat_singleton = &LoadStat{
	userIdMap: &sync.Map{},
}

// 获取负载统计单例对象
func GetLoadStat() *LoadStat {
	return loadStat_singleton
}

func (this *LoadStat) AddUserId(gatewayServerId int32, userId int64) {
	if gatewayServerId <= 0 ||
		userId <= 0 {
		return
	}

	innerMap, ok := this.userIdMap.Load(gatewayServerId)

	if !ok {
		innerMap = &sync.Map{}
		loadedMap, loaded := this.userIdMap.LoadOrStore(gatewayServerId, innerMap)

		if loaded {
			innerMap = loadedMap
		}
	}

	_, loaded := innerMap.(*sync.Map).LoadOrStore(userId, true)

	if !loaded {
		atomic.AddInt32(&this.totalCount, 1)
	}
}

func (this *LoadStat) DeleteUserId(gatewayServerId int32, userId int64) {
	if gatewayServerId <= 0 ||
		userId <= 0 {
		return
	}

	innerMap, ok := this.userIdMap.Load(gatewayServerId)

	if !ok {
		return
	}

	_, loaded := innerMap.(*sync.Map).LoadAndDelete(userId)

	if loaded {
		atomic.AddInt32(&this.totalCount, -1)
	}
}

// 当网关服务器宕机时调用
func (this *LoadStat) DeleteByGatewayServerId(gatewayServerId int32) {
	if gatewayServerId <= 0 {
		return
	}

	val, loaded := this.userIdMap.LoadAndDelete(gatewayServerId)

	if !loaded {
		return
	}

	var deleteCount int32 = 0

	val.(*sync.Map).Range(func(_, _ interface{}) bool {
		deleteCount++
		return true
	})

	atomic.AddInt32(&this.totalCount, -deleteCount)
}

func (this *LoadStat) GetTotalCount() int32 {
	return this.totalCount
}
