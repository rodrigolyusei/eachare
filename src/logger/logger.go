package logger

import (
	"bytes"
	"fmt"
	"log"
)

type LogLevel int

const (
	ZERO LogLevel = iota
	INFO
	DEBUG
	ERROR
)

var infoBuf bytes.Buffer
var infoLogger *log.Logger

var debugBuf bytes.Buffer
var debugLogger *log.Logger

var errorBuf bytes.Buffer
var errorLogger *log.Logger

var logLevel = INFO

func SetLogLevel(level LogLevel) {
	logLevel = level
}

func init() {
	infoLogger = log.New(&infoBuf, "", 0)
	debugLogger = log.New(&debugBuf, "[DEBUG] ", log.Lmicroseconds|log.Lmsgprefix)
	errorLogger = log.New(&errorBuf, "[ERROR] ", log.Lmicroseconds|log.Lmsgprefix|log.Lshortfile)
}

func Info(str string) {
	infoLogger.Output(2, str)
	if logLevel >= INFO {
		fmt.Print(infoBuf.String())
		infoBuf.Reset()
	}
}

func Debug(str string) {
	debugLogger.Output(2, str)
	if logLevel >= DEBUG {
		fmt.Print(debugBuf.String())
		debugBuf.Reset()
	}
}

func Error(str string) {
	errorLogger.Output(2, str)
	if logLevel >= ERROR {
		fmt.Print(errorBuf.String())
		errorBuf.Reset()
	}
}

func (l LogLevel) String() string {
	switch l {
	case ZERO:
		return "ZERO"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
