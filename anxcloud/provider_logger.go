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

func LogDebug(msg string, args ...interface{}) {
	fmt.Fprintf(log.Writer(), debugPrefix+msg, args...)
}

func LogInfo(msg string, args ...interface{}) {
	fmt.Fprintf(log.Writer(), infoPrefix+msg, args...)
}

func LogError(msg string, args ...interface{}) {
	fmt.Fprintf(log.Writer(), errorPrefix+msg, args...)
}

type debugWriter struct {
	writer io.Writer
}

func (w debugWriter) Write(p []byte) (int, error) {
	LogDebug(string(p))
	//msg := append([]byte("[DEBUG]"), p...)
	//return a.writer.Write(msg)
	return 0, nil
}
