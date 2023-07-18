package user_lock

import "sync"

var lockMap = &sync.Map{}

func TryLock(key string) bool {
	if len(key) <= 0 {
		return false
	}

	_, loaded := lockMap.LoadOrStore(key, 1)
	return !loaded
}

func UnLock(key string) {
	if len(key) <= 0 {
		return
	}

	lockMap.Delete(key)
}

func HasLock(key string) bool {
	_, loaded := lockMap.Load(key)
	return loaded
}
