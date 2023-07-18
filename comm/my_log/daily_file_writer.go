package my_log

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

//
// 日志文件写手
//
type dailyFileWriter struct {
	// 文件名称
	fileName string
	// 上一次写入日期
	lastYearDay int
	// 读写锁
	mutex sync.RWMutex
	// 日志队列
	logQ chan []byte
	// 输出文件
	outputFile *os.File
}

// @override
func (w *dailyFileWriter) Write(byteArray []byte) (n int, err error) {
	if nil == byteArray ||
		len(byteArray) <= 0 {
		return 0, nil
	}

	if nil == w.logQ {
		return 0, errors.Errorf("日志项队列为空")
	}

	// 令字节数组入队
	w.logQ <- byteArray

	return len(byteArray), nil
}

// 开始工作, 通过 go 指令来调用
// 也就是异步方式写入日志文件...
func (w *dailyFileWriter) startWork() {
	for log := range w.logQ {
		if nil == log ||
			len(log) <= 0 {
			continue
		}

		// 获取输出文件
		outputFile, err := w.getOutputFile()

		if nil != err ||
			nil == outputFile {
			continue
		}

		// 将日志打印到屏幕和写入文件
		_, _ = os.Stdout.Write(log)
		_, _ = outputFile.Write(log)
	}
}

// 获取输出文件,
// 每天创建一个新的日志文件
func (w *dailyFileWriter) getOutputFile() (io.Writer, error) {
	yearDay := time.Now().YearDay()

	if w.lastYearDay == yearDay &&
		nil != w.outputFile {
		// 如果当前日期和上一次日期一样,
		// 且输出文件也不为空,
		return w.outputFile, nil
	}

	w.lastYearDay = yearDay

	// 构建日志目录
	err := os.MkdirAll(path.Dir(w.fileName), os.ModePerm)

	if nil != err {
		return nil, errors.Wrap(err, "创建目录失败")
	}

	// 定义新的日志文件
	newDailyFile := w.fileName + "." + time.Now().Format("20060102")

	outputFile, err := os.OpenFile(
		newDailyFile,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)

	if nil != err ||
		nil == outputFile {
		return nil, errors.Errorf("打开文件 %s 失败! err = %v", newDailyFile, err)
	}

	if nil != w.outputFile {
		// 关闭原来的文件
		_ = w.outputFile.Close()
	}

	w.outputFile = outputFile
	return outputFile, nil
}
