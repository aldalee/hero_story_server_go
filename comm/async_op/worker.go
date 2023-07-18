package async_op

import (
	"hero_story.go_server/comm/log"
	"hero_story.go_server/comm/main_thread"
)

// 理解为其中的一个线程 + 队列
type worker struct {
	taskQ chan func() // LinkedBlockingQueue<Function> taskQ
}

// 处理异步过程,
// XXX 注意: 这里只是将异步操作加入到队列里, 并不立即执行...
func (w *worker) process(asyncOp func(), continueWith func()) {
	// 这个 w *worker 就相当于 this

	if nil == asyncOp {
		log.Error("异步操作为空")
		return
	}

	if nil == w.taskQ {
		log.Error("任务队列尚未初始化")
		return
	}

	w.taskQ <- func() { // taskQ.offer(new Function() { ... })
		// 执行异步操作
		asyncOp()

		if nil != continueWith {
			// 回到主线程继续执行
			main_thread.Process(continueWith)
		}
	}
}

// 循环执行任务
func (w *worker) loopExecTask() {
	if nil == w.taskQ {
		log.Error("任务队列尚未初始化")
		return
	}

	for {
		task := <-w.taskQ

		if nil == task {
			continue
		}

		func() {
			defer func() {
				if err := recover(); nil != err { // try ... catch ...
					log.Error("发生异常, %+v", err)
				}
			}()

			task()
		}()
	}
}
