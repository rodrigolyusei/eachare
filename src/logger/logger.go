package logger

// Pacotes nativos de go
import (
	"bytes"
	"io"
	"log"
	"os"
	"sync"
)

// Define uma int para o nível do log
type LogLevel uint8

// Define uma enum para os níveis do log
const (
	ZERO LogLevel = iota
	INFO
	DEBUG
	ERROR
)

// Estrutura para armazenar a mensagem de log e seu nível
type LogMessage struct {
	level   LogLevel
	message string
}

// Variáveis globais para o logger
var logQueue = make(chan LogMessage, 100)
var logLevel = INFO
var outputBuf io.Writer

// Variáveis para o buffer da mensagem e logger para escrita
var infoBuf bytes.Buffer
var infoLogger *log.Logger
var debugBuf bytes.Buffer
var debugLogger *log.Logger
var errorBuf bytes.Buffer
var errorLogger *log.Logger

// Mutex para proteger o buffer de saída
var stdBufMutex sync.Mutex
var infoBufMutex sync.Mutex
var debugBufMutex sync.Mutex
var errorBufMutex sync.Mutex

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

// Define a saída padrão para o logger
func SetOutput(w io.Writer) {
	if w == nil {
		outputBuf = os.Stdout
	} else {
		outputBuf = w
	}
}

// Função para ler da fila de logs e escrever no buffer de saída
func ConsumeLogQueue(ch chan LogMessage) {
	for {
		logMessage := <-ch
		outputBuf.Write([]byte(logMessage.message))
	}
}

// init() é chamado na execução automaticamente, e aqui define o padrão pro log
func init() {
	SetOutput(os.Stdout)
	go ConsumeLogQueue(logQueue)

	infoLogger = log.New(&infoBuf, "\t", 0)
	debugLogger = log.New(&debugBuf, "[DEBUG] ", log.Lmicroseconds|log.Lmsgprefix)
	errorLogger = log.New(&errorBuf, "[ERROR] ", log.Lmicroseconds|log.Lmsgprefix|log.Lshortfile)
}

// Funções para logar mensagens de diferentes níveis
func Std(str string) {
	stdBufMutex.Lock()
	defer stdBufMutex.Unlock()

	if logLevel >= ZERO {
		logQueue <- LogMessage{level: ZERO, message: str}
		infoBuf.Reset()
	}
}

func Info(str string) {
	infoBufMutex.Lock()
	defer infoBufMutex.Unlock()

	infoLogger.Output(2, str)
	if logLevel >= INFO {
		logQueue <- LogMessage{level: INFO, message: infoBuf.String()}
		infoBuf.Reset()
	}
}

func Debug(str string) {
	debugBufMutex.Lock()
	defer debugBufMutex.Unlock()

	debugLogger.Output(2, str)
	if logLevel >= DEBUG {
		logQueue <- LogMessage{level: DEBUG, message: infoBuf.String()}
		debugBuf.Reset()
	}
}

func Error(str string) {
	errorBufMutex.Lock()
	defer errorBufMutex.Unlock()

	errorLogger.Output(2, str)
	if logLevel >= ERROR {
		logQueue <- LogMessage{level: ERROR, message: infoBuf.String()}
		errorBuf.Reset()
	}
}
