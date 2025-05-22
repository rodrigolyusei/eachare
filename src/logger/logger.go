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

type Logger struct {
	buffer *bytes.Buffer
	logger *log.Logger
	mutex  sync.Mutex
	Level  LogLevel
}

func (l *Logger) Lock() {
	l.mutex.Lock()
}

func (l *Logger) Unlock() {
	l.mutex.Unlock()
}

func (l *Logger) Write(message string) LogMessage {
	stdLogger.Lock()
	defer stdLogger.Unlock()

	l.logger.Print(message)

	bufMessage := l.buffer.String()
	l.buffer.Reset()

	return LogMessage{level: l.Level, message: bufMessage}
}

// Variáveis globais para o logger
var logQueue = make(chan LogMessage, 100)
var logLevel = INFO
var outputBuf io.Writer

// Variáveis para o buffer da mensagem e logger para escrita
var stdLogger Logger
var infoLogger Logger
var debugLogger Logger
var errorLogger Logger

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

	var stdBuf, infoBuf, debugBuf, errorBuf bytes.Buffer

	stdLogger = Logger{
		buffer: &stdBuf,
		logger: log.New(&stdBuf, "", 0),
	}

	infoLogger = Logger{
		buffer: &infoBuf,
		logger: log.New(&infoBuf, "\t", 0),
		Level:  INFO,
	}

	debugLogger = Logger{
		buffer: &debugBuf,
		logger: log.New(&debugBuf, "[DEBUG] ", log.Lmicroseconds|log.Lmsgprefix),
		Level:  DEBUG,
	}

	errorLogger = Logger{
		buffer: &errorBuf,
		logger: log.New(&errorBuf, "[ERROR] ", log.Lmicroseconds|log.Lmsgprefix|log.Lshortfile),
		Level:  ERROR,
	}
}

// Funções para logar mensagens de diferentes níveis
func Std(str string) {
	if logLevel >= ZERO {
		logQueue <- stdLogger.Write(str)
	}
}

func Info(str string) {
	if logLevel >= INFO {
		logQueue <- infoLogger.Write(str)
	}
}

func Debug(str string) {
	if logLevel >= DEBUG {
		logQueue <- debugLogger.Write(str)
	}
}

func Error(str string) {
	if logLevel >= ERROR {
		logQueue <- errorLogger.Write(str)
	}
}
