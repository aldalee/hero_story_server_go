package my_log

import (
	"fmt"
	"log"
)

// 日志队列大小
const logQSize = 2048

var writer *dailyFileWriter
var debugLogger, infoLogger, warningLogger, errorLogger *log.Logger

//
// Init 初始化
//
func Init(fileName string) {
	writer = &dailyFileWriter{
		fileName:    fileName,
		lastYearDay: -1,
		logQ:        make(chan []byte, logQSize),
	}

	debugLogger = log.New(
		writer, "[ DEBUG ] ",
		log.Ltime|log.Lmicroseconds|log.Lshortfile,
	)

	infoLogger = log.New(
		writer, "[ INFO ] ",
		log.Ltime|log.Lmicroseconds|log.Lshortfile,
	)

	warningLogger = log.New(
		writer, "[ WARNING ] ",
		log.Ltime|log.Lmicroseconds|log.Lshortfile,
	)

	errorLogger = log.New(
		writer, "[ ERROR ] ",
		log.Ltime|log.Lmicroseconds|log.Lshortfile,
	)

	// 写手开始工作
	go writer.startWork()
}

// Debug 记录调试日志
func Debug(format string, valArray ...interface{}) {
	_ = debugLogger.Output(
		2,
		fmt.Sprintf(format, valArray...),
	)
}

// Info 记录消息日志
func Info(format string, valArray ...interface{}) {
	_ = infoLogger.Output(
		2,
		fmt.Sprintf(format, valArray...),
	)
}

// Warning 记录警告日志
func Warning(format string, valArray ...interface{}) {
	_ = warningLogger.Output(
		2,
		fmt.Sprintf(format, valArray...),
	)
}

// Error 记录错误日志
func Error(format string, valArray ...interface{}) {
	_ = errorLogger.Output(
		2,
		fmt.Sprintf(format, valArray...),
	)
}
