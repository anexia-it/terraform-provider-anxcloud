package anxcloud

import (
	"fmt"
	"io"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

// terraformr implements a logr.LogSink for use with Terraform. Terraform expects logged lines to start with
// a defined prefix to specify the log level of the line. It uses funcr.Formatter to build the log messages,
// prefixing them to be terraform compatible.
type terraformr struct {
	funcr.Formatter

	writer io.Writer
}

func (l terraformr) WithName(name string) logr.LogSink {
	l.Formatter.AddName(name)
	return &l
}

func (l terraformr) WithValues(kvList ...interface{}) logr.LogSink {
	l.Formatter.AddValues(kvList)
	return &l
}

func (l terraformr) Info(level int, msg string, keysAndValues ...interface{}) {
	var prefix string

	switch level {
	case 0:
		prefix = "[WARN]"
	case 1:
		prefix = "[INFO]"
	case 2:
		prefix = "[DEBUG]"
	default:
		levelSuffix := ""

		if level > 3 {
			levelSuffix = fmt.Sprintf("+%v", level-3)
		}

		prefix = fmt.Sprintf("[TRACE]%v", levelSuffix)
	}

	fmtPrefix, fmtMsg := l.FormatInfo(level, msg, keysAndValues)

	// ignoring the return value because what should we do on error .. log it?
	_, _ = io.WriteString(l.writer, fmt.Sprintf("%v %v %v", prefix, fmtPrefix, fmtMsg))
}

func (l terraformr) Error(err error, msg string, keysAndValues ...interface{}) {
	fmtPrefix, fmtMsg := l.FormatError(err, msg, keysAndValues)

	// ignoring the return value because what should we do on error .. log it?
	_, _ = io.WriteString(l.writer, fmt.Sprintf("%v %v %v", "[ERROR]", fmtPrefix, fmtMsg))
}

// NewTerraformr creates a logr.Logger logging in terraform compatible format.
func NewTerraformr(w io.Writer) logr.Logger {
	logsink := terraformr{
		writer: w,
		Formatter: funcr.NewFormatter(funcr.Options{
			Verbosity:     4,
			LogCaller:     funcr.Error,
			LogCallerFunc: true,
			LogTimestamp:  true,
		}),
	}
	logsink.AddCallDepth(1)

	return logr.New(&logsink)
}
