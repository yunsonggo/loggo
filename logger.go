package loggo

import (
	"errors"
	"io"
	"os"
	"path"
	"sync"
	"sync/atomic"
)

type Logger interface {
	Error(...interface{})
	ErrorFormat(string, ...interface{})
	Info(...interface{})
	InfoFormat(string, ...interface{})
}

var (
	ErrPath      = errors.New("here need the Config.Path")
	ErrInit      = errors.New("please initialized the loggo before calling")
	ErrNameSpace = errors.New("here need the Config.NameSpace")
	writeConsole bool
	InfoLog      io.WriteCloser
	ErrorLog     io.WriteCloser
	SlowLog      io.WriteCloser
	StateLog     io.WriteCloser
	StackLog     *LessLogger
	once         sync.Once
	initialized  uint32
	options      logOptions
)


func loadConfig(c Config) error {
	if c.LogMode == VARMODE {
		return loadWithVolume(c)
	} else {
		return loadWithFiles(c)
	}
}

func loadWithVolume(c Config) error {
	if len(c.NameSpace) == 0 {
		return ErrNameSpace
	}
	hostname := getHostname()
	c.Path = path.Join(c.Path, c.NameSpace, hostname)
	return loadWithFiles(c)
}

func loadWithFiles(c Config) (err error) {
	var opts []logOption
	if len(c.Path) == 0 {
		return ErrPath
	}
	opts = append(opts, withCoolDownMillis(c.StackCoolDownMillis))
	if c.Compress {
		opts = append(opts, withGzip())
	}
	if c.LastingDays > 0 {
		opts = append(opts, withLastingDays(c.LastingDays))
	}
	accessFile := path.Join(c.Path, AccessFile)
	errorFile := path.Join(c.Path, ErrorFile)
	slowFile := path.Join(c.Path, SlowFile)
	StateFile := path.Join(c.Path, StateFile)

	once.Do(func() {
		handleOptions(opts)
		if InfoLog, err = createOutput(accessFile, InfoPrefix, c.Stdout); err != nil {
			return
		}
		if ErrorLog, err = createOutput(errorFile, ErrorPrefix, c.Stdout); err != nil {
			return
		}
		if SlowLog, err = createOutput(slowFile, SlowPrefix, c.Stdout); err != nil {
			return
		}
		if StateLog, err = createOutput(StateFile, StatePrefix, c.Stdout); err != nil {
			return
		}
		StackLog = NewLessLogger(options.logStackCoolDownMills)
		atomic.StoreUint32(&initialized,1)
	})
	return err
}

func createOutput(path, prefix string, stdout bool) (io.WriteCloser, error) {
	if len(path) == 0 {
		return nil, ErrPath
	}
	return NewLogger(path, stdout, defaultRule(path, prefix, FileDelimiter, options.lastingDays, options.gzipEnabled), options.gzipEnabled)
}

func withCoolDownMillis(millis int) logOption {
	return func(opts *logOptions) {
		opts.logStackCoolDownMills = millis
	}
}

func withGzip() logOption {
	return func(opts *logOptions) {
		opts.gzipEnabled = true
	}
}

func withLastingDays(days int) logOption {
	return func(opts *logOptions) {
		opts.lastingDays = days
	}
}

// find host name
// will use default host name if not found
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil || len(hostname) == 0 {
		return DefaultHostname
	}
	return hostname
}
