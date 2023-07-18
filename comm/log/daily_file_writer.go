package log

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

type dailyFileWriter struct {
	// 日志文件名称
	fileName string
	// 上一次写入日期
	lastYearDay int
	// 输出文件
	outputFile *os.File
	// 文件交换锁
	fileSwitchLock *sync.Mutex
}

// Write 输出日志
func (w *dailyFileWriter) Write(byteArray []byte) (n int, err error) {
	if nil == byteArray ||
		len(byteArray) <= 0 {
		return 0, nil
	}

	outputFile, err := w.getOutputFile()

	if nil != err {
		return 0, err
	}

	_, _ = os.Stderr.Write(byteArray)
	_, _ = outputFile.Write(byteArray)

	return len(byteArray), nil
}

// 获取输出文件
// 每天创建一个新得日志文件
func (w *dailyFileWriter) getOutputFile() (io.Writer, error) {
	yearDay := time.Now().YearDay()

	if w.lastYearDay == yearDay &&
		nil != w.outputFile {
		return w.outputFile, nil
	}

	w.fileSwitchLock.Lock()
	defer w.fileSwitchLock.Unlock()

	if w.lastYearDay == yearDay &&
		nil != w.outputFile {
		return w.outputFile, nil
	}

	w.lastYearDay = yearDay

	// 先建立日志目录
	err := os.MkdirAll(path.Dir(w.fileName), os.ModePerm)

	if nil != err {
		return nil, err
	}

	// 定义日志文件名称 = 日志文件名 . 日期后缀
	newDailyFile := w.fileName + "." + time.Now().Format("20060102")

	outputFile, err := os.OpenFile(
		newDailyFile,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644, // rw-r--r--
	)

	if nil != err ||
		nil == outputFile {
		return nil, errors.Errorf("打开文件 %s 失败, err = %v", newDailyFile, err)
	}

	if nil != w.outputFile {
		// 关闭原来的文件
		_ = w.outputFile.Close()
	}

	w.outputFile = outputFile
	return outputFile, nil
}
