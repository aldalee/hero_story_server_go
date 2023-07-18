package lazy_save

import (
	"hero_story.go_server/comm/log"
	"sync"
	"time"
)

var lsoMap = &sync.Map{}

func init() {
	startSave()
}

func SaveOrUpdate(lso LazySaveObj) {
	if nil == lso {
		return
	}

	log.Info("记录延迟保存对象, lsoId = %s", lso.GetLsoId())

	nowTime := time.Now().UnixMilli()
	existRecord, _ := lsoMap.Load(lso.GetLsoId())

	if nil != existRecord {
		existRecord.(*lazySaveRecord).setLastUpdateTime(nowTime)
		return
	}

	newRecord := &lazySaveRecord{}
	newRecord.lsoRef = lso
	newRecord.setLastUpdateTime(nowTime)
	lsoMap.Store(lso.GetLsoId(), newRecord)
}

func startSave() {
	go func() {
		for {
			time.Sleep(time.Second)

			nowTime := time.Now().UnixMilli()
			deleteLsoIdArray := make([]string, 64)

			lsoMap.Range(func(_, val interface{}) bool { // for (Map.Entry entry : mapObj)
				if nil == val {
					return true
				}

				currRecord := val.(*lazySaveRecord)

				if nowTime-currRecord.getLastUpdateTime() < 20000 {
					// 如果延迟保存对象的最后更新时间和当前时间下差不过 20 秒,
					// 等等再说吧...
					return true
				}

				log.Info("执行延迟保存, lsoId = %s", currRecord.lsoRef.GetLsoId())

				// 执行保存逻辑
				currRecord.lsoRef.SaveOrUpdate(nil)

				deleteLsoIdArray = append(deleteLsoIdArray, currRecord.lsoRef.GetLsoId())
				return true
			})

			for _, lsoId := range deleteLsoIdArray {
				lsoMap.Delete(lsoId)
			}
		}
	}()
}
