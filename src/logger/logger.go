package logger

// Pacotes nativos de go
import (
	"bytes"
	"fmt"
	"log"
)

// Define uma int para o nível de log
type LogLevel int

// Define uma enum para os níveis
const (
	ZERO LogLevel = iota
	INFO
	DEBUG
	ERROR
)

// Variáveis para o buffer da mensagem e logger para escrita
var infoBuf bytes.Buffer
var infoLogger *log.Logger

var debugBuf bytes.Buffer
var debugLogger *log.Logger

var errorBuf bytes.Buffer
var errorLogger *log.Logger

// logLevel recebe 1 de ínicio
var logLevel = INFO

// Setter para o logLevel
func SetLogLevel(level LogLevel) {
	logLevel = level
}

// Retorna o logLevel atual como string
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

// init() é chamado na execução automaticamente, e aqui define o padrão pro log
func init() {
	infoLogger = log.New(&infoBuf, "", 0)
	debugLogger = log.New(&debugBuf, "[DEBUG] ", log.Lmicroseconds|log.Lmsgprefix)
	errorLogger = log.New(&errorBuf, "[ERROR] ", log.Lmicroseconds|log.Lmsgprefix|log.Lshortfile)
}

// Funções a serem usadas no programa para inmprimir cada tipo de mensagem
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
