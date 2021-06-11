package anxcloud

import (
	"fmt"
	"io"
	"log"
)

const (
	debugPrefix = "[DEBUG]"
	infoPrefix  = "[INFO]"
	errorPrefix = "[ERROR]"
)

func Debug(i ...interface{}) {
	fmt.Fprint(log.Writer(), debugPrefix, i)
}

func Info(i ...interface{}) {
	fmt.Fprint(log.Writer(), infoPrefix, i)
}

func Error(i ...interface{}) {
	fmt.Fprint(log.Writer(), errorPrefix, i)
}


type AnxLogger struct {
	prefix string
	writer io.Writer
}

func (a AnxLogger) Write(p []byte) (int, error) {
	msg := append([]byte(a.prefix), p...)
	return a.writer.Write(msg)
}

func NewAnxLogger(prefix string, writer io.Writer) AnxLogger {
	return AnxLogger{
		prefix: prefix,
		writer: writer,
	}
}
