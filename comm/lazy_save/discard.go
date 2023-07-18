package lazy_save

import "hero_story.go_server/comm/log"

func Discard(lso LazySaveObj) {
	if nil == lso {
		return
	}

	log.Info("放弃延迟保存, lsoId = %+v", lso.GetLsoId())

	lsoMap.Delete(lso.GetLsoId())
}
