package logger

import (
	"fmt"
	"log"
	"os"
)

const LogFile = "/tmp/shell.log"

func newLogger() Logger {
	return &logger{}
}

type logger struct {
}

func (l *logger) WriteMessages(msg ...interface{}) {
	file, err := os.OpenFile(LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)

	if err != nil {
		log.Fatal(err.Error())
	}

	defer file.Close()

	_, _ = file.WriteString(fmt.Sprintln(msg...))
}
