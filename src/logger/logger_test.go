package logger

import (
	"bytes"
	"os"
	"regexp"
	"testing"
)

func changeStdout(t *testing.T, level LogLevel, logFunc func(str string)) bytes.Buffer {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Substitui a saída padrão por `w`
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }() // Restaura a saída padrão ao final do teste

	// Define o nível de log
	SetLogLevel(level)

	// Chama a função Info
	logFunc("Hello world!")

	// Fecha o writer para finalizar a escrita no pipe
	w.Close()

	// Lê a saída capturada
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		panic(err)
	}
	return buf
}

func TestInfoLog(t *testing.T) {
	Info("Hello world!")
	if infoBuf.String() != "Hello world!\n" {
		t.Errorf("Expected 'Hello world!', got '%s'", infoBuf.String())
	}
}

func TestDebugLog(t *testing.T) {
	Debug("Hello world!")
	regex := `^\d{2}:\d{2}:\d{2}\.\d{6} \[DEBUG\] Hello world!\n`

	matched, err := regexp.MatchString(regex, debugBuf.String())
	if err != nil {
		t.Fatalf("Error matching regex: %v", err)
	}
	if !matched {
		t.Errorf("Log message '%s' does not match expected format", debugBuf.String())
	}
}

func TestErrorLog(t *testing.T) {
	Error("Hello world!")
	regex := `^\d{2}:\d{2}:\d{2}\.\d{6} logger_test.go:29: \[ERROR\] Hello world!\n`

	matched, err := regexp.MatchString(regex, errorBuf.String())
	if err != nil {
		t.Fatalf("Error matching regex: %v", err)
	}
	if !matched {
		t.Errorf("Log message '%s' does not match expected format", errorBuf.String())
	}
}

func TestInfoOutputOk(t *testing.T) {
	logLevels := []LogLevel{INFO, DEBUG, ERROR}
	for _, level := range logLevels {
		t.Run(level.String(), func(t *testing.T) {
			buf := changeStdout(t, level, Info)

			output := buf.String()
			expected := "Hello world!\n"
			if output != expected {
				t.Errorf("Expected '%s', got '%s' for log level '%s'", expected, output, level.String())
			}
		})
	}
}

func TestInfoOutputWrong(t *testing.T) {

	buf := changeStdout(t, ZERO, Info)

	output := buf.String()
	expected := ""
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}

func TestDebugOutputOk(t *testing.T) {
	logLevels := []LogLevel{DEBUG, ERROR}
	for _, level := range logLevels {
		t.Run(level.String(), func(t *testing.T) {

			buf := changeStdout(t, level, Debug)

			output := buf.String()
			regex := `^\d{2}:\d{2}:\d{2}\.\d{6} \[DEBUG\] Hello world!\n`
			matched, err := regexp.MatchString(regex, output)
			if err != nil {
				t.Fatalf("Error matching regex: %v", err)
			}
			if !matched {
				t.Errorf("Log message '%s' does not match expected format", output)
			}
		})
	}
}

func TestDebugOutputError(t *testing.T) {
	logLevels := []LogLevel{ZERO, INFO}
	for _, level := range logLevels {
		t.Run(level.String(), func(t *testing.T) {
			buf := changeStdout(t, level, Debug)

			output := buf.String()
			expected := ""
			if output != expected {
				t.Errorf("Log message '%s' does not match expected format", output)
			}
		})
	}
}

func TestErrorOutputOk(t *testing.T) {
	logLevels := []LogLevel{ERROR}
	for _, level := range logLevels {
		t.Run(level.String(), func(t *testing.T) {
			buf := changeStdout(t, level, Error)

			output := buf.String()
			regex := `^\d{2}:\d{2}:\d{2}\.\d{6} logger_test.go:25: \[ERROR\] Hello world!\n`
			matched, err := regexp.MatchString(regex, output)
			if err != nil {
				t.Fatalf("Error matching regex: %v", err)
			}
			if !matched {
				t.Errorf("Log message '%s' does not match expected format", output)
			}
		})
	}
}

func TestErrorOutputError(t *testing.T) {
	logLevels := []LogLevel{ZERO, INFO, DEBUG}
	for _, level := range logLevels {
		t.Run(level.String(), func(t *testing.T) {
			buf := changeStdout(t, level, Error)

			output := buf.String()
			expected := ""
			if output != expected {
				t.Errorf("Log message '%s' does not match expected format", output)
			}
		})
	}
}
