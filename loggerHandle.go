package loggo

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func Error(v ...interface{}) {
	errorCaller(1, v...)
}

func ErrorFormat(format string, v ...interface{}) {
	errorCallerFormat(1, format, v...)
}

func errorCaller(callDepth int, v ...interface{}) {
	errorSync(fmt.Sprintln(v...), callDepth+CallerInnerdepth)
}

func errorCallerFormat(callDepth int, format string, v ...interface{}) {
	errorSync(fmt.Sprintf(fmt.Sprintf("%s\n", format), v...), callDepth+CallerInnerdepth)
}

func Info(v ...interface{}) {
	infoSync(fmt.Sprintln(v...))
}

func InfoFormat(format string, v ...interface{}) {
	infoSync(fmt.Sprintf(fmt.Sprintf("%s\n", format), v...))
}

func server(v ...interface{}) {
	stackSync(fmt.Sprint(v...))
}

func ServerFormat(format string, v ...interface{}) {
	stackSync(fmt.Sprintf(format, v...))
}

func Slow(v ...interface{}) {
	slowSync(fmt.Sprintln(v...))
}

func SlowFormat(format string, v ...interface{}) {
	slowSync(fmt.Sprintf(fmt.Sprintf("%s\n", format), v...))
}

func Stat(v ...interface{}) {
	statSync(fmt.Sprintln(v...))
}

func StatFormat(format string, v ...interface{}) {
	statSync(fmt.Sprintf(fmt.Sprintf("%s\n", format), v...))
}

func errorSync(msg string, callDepth int) {
	if atomic.LoadUint32(&initialized) == 0 {
		stdoutErrOutput(ErrorPrefix, msg, callDepth)
	} else {
		outputError(ErrorLog, msg, callDepth)
	}
}

func infoSync(msg string) {
	if atomic.LoadUint32(&initialized) == 0 {
		stdoutOutput(InfoPrefix, msg)
	} else {
		output(InfoLog, msg)
	}
}

func stackSync(msg string) {
	if atomic.LoadUint32(&initialized) == 0 {
		stdoutOutput(StackPrefix, fmt.Sprintf("%s\n%s", msg, string(debug.Stack())))
	} else {
		StackLog.Errorf("%s\n%s", msg, string(debug.Stack()))
	}
}

func slowSync(msg string) {
	if atomic.LoadUint32(&initialized) == 0 {
		stdoutOutput(SlowPrefix, msg)
	} else {
		output(SlowLog, msg)
	}
}

func statSync(msg string) {
	if atomic.LoadUint32(&initialized) == 0 {
		stdoutOutput(StatePrefix, msg)
	} else {
		output(StateLog, msg)
	}
}

func stdoutOutput(prefix, msg string) {
	fmt.Print(prefix + addTime(msg))
}

func stdoutErrOutput(prefix, msg string, callDepth int) {
	fmt.Print(prefix + addTimeAndCaller(msg, callDepth))
}

func getCaller(callDepth int) string {
	var buf strings.Builder
	_, file, line, ok := runtime.Caller(callDepth)
	if ok {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		buf.WriteString(short)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(line))
	}
	return buf.String()
}

func output(writer io.Writer, msg string) {
	buf := addTime(msg)
	if writer != nil {
		if _, err := writer.Write([]byte(buf)); err != nil {
			log.Println(err)
		}
	}
}

func outputError(writer io.Writer, msg string, callDepth int) {
	content := addTimeAndCaller(msg, callDepth)
	if writer != nil {
		if _, err := writer.Write([]byte(content)); err != nil {
			log.Println(err)
		}
	}
}

func addTime(msg string) string {
	now := []byte(time.Now().Format(TimeFormat))
	msgBytes := []byte(msg)
	buf := make([]byte, len(now)+1+len(msgBytes))
	n := copy(buf, now)
	buf[n] = ' '
	copy(buf[n+1:], msgBytes)
	return string(buf)
}

func addTimeAndCaller(msg string, callDepth int) string {
	var buf strings.Builder
	buf.WriteString(time.Now().Format(TimeFormat))
	buf.WriteByte(' ')
	caller := getCaller(callDepth)
	if len(caller) > 0 {
		buf.WriteString(caller)
		buf.WriteByte(' ')
	}
	buf.WriteString(msg)
	return buf.String()
}

func Close() error {
	if writeConsole {
		return nil
	}
	if atomic.LoadUint32(&initialized) == 0 {
		return ErrInit
	}
	atomic.StoreUint32(&initialized,0)
	if InfoLog != nil {
		if err := InfoLog.Close(); err != nil {
			return err
		}
	}
	if ErrorLog != nil {
		if err := ErrorLog.Close(); err != nil {
			return err
		}
	}
	if SlowLog != nil {
		if err := SlowLog.Close(); err != nil {
			return err
		}
	}

	if StateLog != nil {
		if err := StateLog.Close(); err != nil {
			return err
		}
	}

	return nil
}

