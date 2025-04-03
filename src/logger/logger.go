package logger

// Pacotes nativos de go
import (
	"bytes"
	"io"
	"log"
	"os"
)

type LogLevel int

const (
	ZERO LogLevel = iota
	INFO
	DEBUG
	ERROR
)

type LogMessage struct {
	Level   LogLevel
	Message string
}

var infoBuf bytes.Buffer
var infoLogger *log.Logger

var debugBuf bytes.Buffer
var debugLogger *log.Logger

var errorBuf bytes.Buffer
var errorLogger *log.Logger

var outputBuf io.Writer

var logQueue = make(chan LogMessage, 100)

var logLevel = INFO

func ConsumeLogQueue(ch chan LogMessage) {
	for {
		logMessage := <-ch
		outputBuf.Write([]byte(logMessage.Message))
	}
}

func SetLogLevel(level LogLevel) {
	logLevel = level
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

func init() {
	SetOutput(os.Stdout)

	go ConsumeLogQueue(logQueue)

	infoLogger = log.New(&infoBuf, "", 0)
	debugLogger = log.New(&debugBuf, "[DEBUG] ", log.Lmicroseconds|log.Lmsgprefix)
	errorLogger = log.New(&errorBuf, "[ERROR] ", log.Lmicroseconds|log.Lmsgprefix|log.Lshortfile)
}

func SetOutput(w io.Writer) {
	if w == nil {
		outputBuf = os.Stdout
	} else {
		outputBuf = w
	}
}

func Info(str string) {
	infoLogger.Output(2, str)
	if logLevel >= INFO {
		//fmt.Println(infoBuf.String())
		logQueue <- LogMessage{Level: INFO, Message: infoBuf.String()}
		infoBuf.Reset()
	}
}

func Debug(str string) {
	debugLogger.Output(2, str)
	if logLevel >= DEBUG {
		//outputBuf.Write(debugBuf.Bytes())
		logQueue <- LogMessage{Level: DEBUG, Message: debugBuf.String()}
		debugBuf.Reset()
	}
}

func Error(str string) {
	errorLogger.Output(2, str)
	if logLevel >= ERROR {
		//outputBuf.Write(errorBuf.Bytes())
		logQueue <- LogMessage{Level: ERROR, Message: errorBuf.String()}
		errorBuf.Reset()
	}
}
