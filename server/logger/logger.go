package logger

import (
    "io"
    "log"
)

var (
    Info  *log.Logger
    Warning *log.Logger
    Error *log.Logger
)

func Init(flags int, out io.Writer) {
    Info = log.New(out, "INFO: ", flags)
    Warning = log.New(out, "WARNING: ", flags)
    Error = log.New(out, "ERROR: ", flags)
}

