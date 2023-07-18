package log

import (
	"fmt"
	"log"
	"sync"
)

var writer *dailyFileWriter
var infoLogger, errorLogger *log.Logger

// Config 配置日志
func Config(outputFileName string) {
	if len(outputFileName) <= 0 {
		panic("输出文件名为空")
	}

	writer = &dailyFileWriter{
		fileName:       outputFileName,
		lastYearDay:    -1,
		fileSwitchLock: &sync.Mutex{},
	}

	infoLogger = log.New(
		writer, "[ INFO ] ",
		log.Ltime|log.Lmicroseconds|log.Lshortfile,
	)

	errorLogger = log.New(
		writer, "[ ERROR ] ",
		log.Ltime|log.Lmicroseconds|log.Lshortfile,
	)
}

// Info 输出消息日志
func Info(format string, valArray ...interface{}) {
	_ = infoLogger.Output(
		2,
		fmt.Sprintf(format, valArray...),
	)
}

// Error 输出错误日志
func Error(format string, valArray ...interface{}) {
	_ = errorLogger.Output(
		2,
		fmt.Sprintf(format, valArray...),
	)
}
