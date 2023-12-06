package mysqlslog

import (
	"fmt"
	"log/slog"
)

type Logger struct {
	Logger *slog.Logger
}

func (l *Logger) Print(v ...interface{}) {
	line := fmt.Sprintln(v...)
	l.Logger.With("subsystem", "mysql").Debug(line)
}
