package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	for _, w := range hook.Writer {
		w.Write([]byte(line))
	}
	return err
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

var logger *logrus.Entry

type Logger struct {
	*logrus.Entry
}

func NewLogger() *Logger {
	return &Logger{logger}
}

func init() {
	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
		FullTimestamp: true,
	}

	file, err := os.Create("all.log")
	if err != nil {
		panic(err)
	}

	l.SetOutput(io.Discard)

	l.AddHook(&writerHook{
		Writer: []io.Writer{
			file,
			os.Stdout},
		LogLevels: logrus.AllLevels,
	})

	l.SetLevel(logrus.TraceLevel)

	logger = logrus.NewEntry(l)
}
