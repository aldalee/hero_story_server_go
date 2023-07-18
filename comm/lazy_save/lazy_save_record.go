package lazy_save

type lazySaveRecord struct {
	lsoRef         LazySaveObj
	lastUpdateTime int64
}

func (record *lazySaveRecord) getLastUpdateTime() int64 {
	return record.lastUpdateTime
}

func (record *lazySaveRecord) setLastUpdateTime(val int64) {
	record.lastUpdateTime = val
}
