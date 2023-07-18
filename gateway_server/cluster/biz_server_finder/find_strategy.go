package biz_server_finder

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"

	"hero_story.go_server/comm/cluster"
	"hero_story.go_server/comm/log"
)

// FindStrategy 查找服务器策略接口
type FindStrategy interface {
	// 给定业务服务器实例字典, 执行查找过程
	doFind(bizServerInstanceMap *sync.Map) (*BizServerInstance, error)
}

// RandomFindStrategy 随机查找策略
type RandomFindStrategy struct {
}

func (finder *RandomFindStrategy) doFind(bizServerInstanceMap *sync.Map) (*BizServerInstance, error) {
	if nil == bizServerInstanceMap {
		return nil, errors.New("参数为空")
	}

	// 用于收集服务器实例的数组
	valArray := make([]interface{}, 0, 10)

	bizServerInstanceMap.Range(func(_, val interface{}) bool {
		valArray = append(valArray, val)
		return true
	})

	count := len(valArray)

	if count <= 0 {
		return nil, errors.New("业务服务器实例数为 0")
	}

	// 随机一个数, 取值范围 = [0, count)
	randIndex := rand.Int31n(int32(count))
	findVal := valArray[randIndex]

	return findVal.(*BizServerInstance), nil
}

// 通过 Etcd 来查找服务器
type EtcdFindStrategy struct {
	EtcdKey       string
	WriteServerId *int32
}

func (this *EtcdFindStrategy) doFind(bizServerInstanceMap *sync.Map) (*BizServerInstance, error) {
	if nil == bizServerInstanceMap {
		return nil, errors.New("参数为空")
	}

	getResp, err := cluster.GetEtcdCli().Get(context.TODO(), this.EtcdKey)

	if nil != err {
		return nil, err
	}

	if len(getResp.Kvs) <= 0 {
		return nil, fmt.Errorf("Etcd 中未找到关键字 '%s'", this.EtcdKey)
	}

	strVal := string(getResp.Kvs[0].Value)
	intVal, err := strconv.Atoi(strVal)

	if nil != err {
		return nil, err
	}

	log.Info("EtcdFindStrategy 找到 Id = %d 的业务服务器", intVal)

	idFindStrategy := &IdFindStrategy{
		ServerId: int32(intVal),
	}

	bizServerInstance, err := idFindStrategy.doFind(bizServerInstanceMap)

	if nil != err {
		return nil, err
	}

	*this.WriteServerId = bizServerInstance.ServerId
	return bizServerInstance, nil
}

type IdFindStrategy struct {
	ServerId int32
}

func (finder *IdFindStrategy) doFind(bizServerInstanceMap *sync.Map) (*BizServerInstance, error) {
	if nil == bizServerInstanceMap {
		return nil, errors.New("参数为空")
	}

	val, ok := bizServerInstanceMap.Load(finder.ServerId)

	if !ok {
		return nil, errors.New(fmt.Sprintf(
			"Id = %d 的服务器不存在",
			finder.ServerId,
		))
	}

	log.Info(
		"IdFindStrategy 找到了 Id = %d 的业务服务器",
		finder.ServerId,
	)

	return val.(*BizServerInstance), nil
}

type LeastLoadFindStrategy struct {
	WriteServerId *int32
}

func (finder *LeastLoadFindStrategy) doFind(bizServerInstanceMap *sync.Map) (*BizServerInstance, error) {
	if nil == bizServerInstanceMap {
		return nil, errors.New("参数为空")
	}

	var findBizServerInstance *BizServerInstance = nil
	var minLoadCount int32 = 999999

	bizServerInstanceMap.Range(func(_, val interface{}) bool {
		if nil == val {
			return true
		}

		currBizServerInstance := val.(*BizServerInstance)

		if currBizServerInstance.LoadCount < minLoadCount {
			findBizServerInstance = currBizServerInstance
			minLoadCount = currBizServerInstance.LoadCount
		}

		return true
	})

	if nil == findBizServerInstance {
		return nil, errors.New("未找到负载最小的服务器")
	}

	log.Info(
		"LeastLoadFindStrategy 找到了 Id = %d 的业务服务器",
		findBizServerInstance.ServerId,
	)

	*finder.WriteServerId = findBizServerInstance.ServerId
	return findBizServerInstance, nil
}

type CompositeFindStrategy struct {
	Finder1 FindStrategy
	Finder2 FindStrategy
	Finder3 FindStrategy
}

func (finder *CompositeFindStrategy) doFind(bizServerInstanceMap *sync.Map) (*BizServerInstance, error) {
	if nil == bizServerInstanceMap {
		return nil, errors.New("参数为空")
	}

	findStrategyArray := [...]FindStrategy{
		finder.Finder1,
		finder.Finder2,
		finder.Finder3,
	}

	for _, findStrategy := range findStrategyArray {
		if nil == findStrategy {
			continue
		}

		bizServerInstance, err := findStrategy.doFind(bizServerInstanceMap)

		if nil == err {
			return bizServerInstance, nil
		}
	}

	// 如果所有策略都没找到,
	// 直接返回错误
	return nil, errors.New("未找到服务器实例")
}
