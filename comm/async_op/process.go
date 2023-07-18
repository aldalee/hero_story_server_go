package async_op

import "sync"

// 工人数组
var workerArray = [2048]*worker{}

// 初始化工人用的锁
var initWorkerLocker = &sync.Mutex{}

// Process 处理异步过程,
// asyncOp 异步函数, 将被放到一个新协程里去执行...
// continueWith 则是回到主线程继续执行的函数
func Process(bindId int, asyncOp func(), continueWith func()) { // asyncOp 和 continueWith 可以理解为 Java 中的 Function
	if nil == asyncOp {
		return
	}

	currWorker := getCurrWorker(bindId)

	if nil != currWorker {
		currWorker.process(asyncOp, continueWith)
	}
}

// 根据 bindId 获取一个工人
func getCurrWorker(bindId int) *worker {
	if bindId < 0 {
		bindId = -bindId
	}

	// 根据数组索引拿到一个干活的工人
	workerIndex := bindId % len(workerArray)
	currWorker := workerArray[workerIndex]

	if nil != currWorker {
		return currWorker
	}

	// 如果工人为空,
	// 则需要初始化这个工人...
	// 这里需要加锁!
	initWorkerLocker.Lock()
	defer initWorkerLocker.Unlock()

	// 重新拿到这个干活的工人
	currWorker = workerArray[workerIndex]

	// 并重新进行空指针判断
	if nil != currWorker {
		return currWorker
	}

	// 双重检查之后这个工人还是空值,
	// 初始化这个工人...
	currWorker = &worker{
		taskQ: make(chan func(), 2048), // 相当于: taskQ = new LinkedBlockingQueue<Function>();
	}

	workerArray[workerIndex] = currWorker
	go currWorker.loopExecTask() // 协程方式开始干活

	return currWorker
}
