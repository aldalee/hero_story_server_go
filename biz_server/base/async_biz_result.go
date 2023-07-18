package base

import (
	"hero_story.go_server/comm/main_thread"
	"sync/atomic"
)

type AsyncBizResult struct {
	// 已返回对象
	returnedObj interface{} // Object
	// 完成回调函数
	completeFunc func()

	// 是否已有返回值
	hasReturnedObj int32
	// 是否已有回调函数
	hasCompleteFunc int32
	// 是否已经调用过完成函数
	completeFuncHasAlreadyBeenCalled int32 // 默认值 = 0, 没被调用过
}

// GetReturnedObj 获取返回值
func (bizResult *AsyncBizResult) GetReturnedObj() interface{} {
	return bizResult.returnedObj
}

// SetReturnedObj 设置返回值
func (bizResult *AsyncBizResult) SetReturnedObj(val interface{}) {
	if atomic.CompareAndSwapInt32(&bizResult.hasReturnedObj, 0, 1) {
		bizResult.returnedObj = val
		bizResult.doComplete()
	}
}

// OnComplete 完成回调函数
func (bizResult *AsyncBizResult) OnComplete(val func()) {
	if atomic.CompareAndSwapInt32(&bizResult.hasCompleteFunc, 0, 1) {
		bizResult.completeFunc = val

		if 1 == bizResult.hasReturnedObj {
			bizResult.doComplete()
		}
	}
}

// 执行完成回调
func (bizResult *AsyncBizResult) doComplete() {
	if nil == bizResult.completeFunc {
		return
	}

	// 通过 CAS 原语来比较标记值,
	// 正确之后才进行调用
	if atomic.CompareAndSwapInt32(&bizResult.completeFuncHasAlreadyBeenCalled, 0, 1) {
		// 扔到主线程里去执行
		main_thread.Process(bizResult.completeFunc)
	}
}
