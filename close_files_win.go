package loggo

import (
	"os"
	"syscall"
)

func closeOnExec(file *os.File) {
	if file != nil {
		syscall.CloseOnExec(syscall.Handle(file.Fd()))
	}
}
