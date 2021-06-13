package anxcloud

import (
	"fmt"
	"io"
	"log"
)

const (
	debugPrefix = "[DEBUG] "
	infoPrefix  = "[INFO]"
	errorPrefix = "[ERROR]"
)

func Debug(i ...interface{}) {
	fmt.Fprint(log.Writer(), append([]interface{}{debugPrefix}, i...)...)
}

func Info(i ...interface{}) {
	fmt.Fprint(log.Writer(), append([]interface{}{infoPrefix}, i...)...)
}

func Error(i ...interface{}) {
	fmt.Fprint(log.Writer(), append([]interface{}{errorPrefix}, i...)...)
}

type debugWriter struct {
	writer io.Writer
}

func (a debugWriter) Write(p []byte) (int, error) {
	msg := append([]byte("[DEBUG]"), p...)
	return a.writer.Write(msg)
}
