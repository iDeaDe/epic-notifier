package utils

import (
	"io"
	"log"
	"os"
)

type Logger log.Logger

func getWriter() io.Writer {
	return os.Stdout
}

func getLogger(prefix string) *log.Logger {
	logger := log.New(getWriter(), prefix, log.LstdFlags)
	return logger
}

func NewLogger(prefix string) *log.Logger {
	return getLogger(prefix)
}
