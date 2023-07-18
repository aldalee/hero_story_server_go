package cluster

import (
	"context"
	"time"

	"go.etcd.io/etcd/client/v3/concurrency"
)

type DLock struct {
	session *concurrency.Session
	mutex   *concurrency.Mutex
}

// 加锁
func NewDLock(lockerName string, maxDuration int64) *DLock {
	if len(lockerName) <= 0 ||
		maxDuration <= 0 {
		return nil
	}

	session, err := concurrency.NewSession(GetEtcdCli())

	if nil != err {
		return nil
	}

	mutex := concurrency.NewMutex(session, lockerName)

	if err := mutex.Lock(context.TODO()); nil != err {
		session.Close()
		return nil
	}

	timer := time.NewTimer(time.Duration(maxDuration) * time.Second)

	newDLock := &DLock{
		session: session,
		mutex:   mutex,
	}

	go func() { // chan == channel 信道
		<-timer.C // chan = channel = java:LinkedBlockingQueue
		// <-LinkedBlockingQueue = LikedBlockingQueue.leftPop 出队
		newDLock.Unlock()
	}()

	return newDLock
}

// 解锁
func (this *DLock) Unlock() {
	if nil != this.mutex {
		this.mutex.Unlock(context.TODO())
	}

	if nil != this.session {
		this.session.Close()
	}
}
