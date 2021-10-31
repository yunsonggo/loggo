package loggo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
)

const (
	DateFormat  = "2006-01-02"
	HoursPerday = 24
	BufferSize  = 100
	DirMode     = 0755
	FileMode    = 0600
)

var ErrFileClosed = errors.New("error:log file was closed")

type LoggerServer struct {
	filename    string
	backup      string
	fp          *os.File
	channel     chan []byte
	done        chan placeHolder
	rule        loggerRule
	compress    bool
	stdout      bool
	lastingDays int
	waitGroup   sync.WaitGroup
	closeOnce   sync.Once
}

func NewLogger(filename string, stdout bool, rule loggerRule, compress bool) (*LoggerServer, error) {
	s := &LoggerServer{
		filename: filename,
		channel:  make(chan []byte, BufferSize),
		done:     make(chan placeHolder),
		rule:     rule,
		stdout:   stdout,
		compress: compress,
	}
	if err := s.init(); err != nil {
		return nil, err
	}
	s.startWorker()
	return s, nil
}

func (ls *LoggerServer) Close() error {
	var err error
	ls.closeOnce.Do(func() {
		close(ls.done)
		ls.waitGroup.Wait()
		if err = ls.fp.Sync(); err != nil {
			return
		}
		err = ls.fp.Close()
	})
	return err
}

func (ls *LoggerServer) Write(data []byte) (int, error) {
	logBytes := append([]byte(ls.rule.getPrefix()), data...)
	select {
	case ls.channel <- logBytes:
		if ls.stdout {
			fmt.Print(string(logBytes))
		}
		return len(logBytes), nil
	case <-ls.done:
		stdoutErrOutput(ls.rule.getPrefix(), string(logBytes), 1)
		return 0, ErrFileClosed
	}
}

func (ls *LoggerServer) getBackupFilename() string {
	if len(ls.backup) == 0 {
		return ls.rule.backupFilename()
	} else {
		return ls.backup
	}
}

func (ls *LoggerServer) init() error {
	ls.backup = ls.rule.backupFilename()
	if _, err := os.Stat(ls.filename); err != nil {
		basePath := path.Dir(ls.filename)
		if _, err = os.Stat(basePath); err != nil {
			if err = os.MkdirAll(basePath, DirMode); err != nil {
				return err
			}
		}
		if ls.fp, err = os.Create(ls.filename); err != nil {
			return err
		}
	} else if ls.fp, err = os.OpenFile(ls.filename, os.O_APPEND|os.O_WRONLY, FileMode); err != nil {
		return err
	}
	closeOnExec(ls.fp)
	return nil
}

func (ls *LoggerServer) maybeCompressFile(file string) {
	if ls.compress {
		defer func() {
			if r := recover(); r != nil {
				server(r)
			}
		}()
		compressLogFile(file)
	}
}

func (ls *LoggerServer) maybeDeleteOutdatedFiles() {
	files := ls.rule.outdatedFiles()
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			ErrorFormat("failed to remove outdated file: %s", file)
		}
	}
}

func (ls *LoggerServer) postRotate(file string) {
	go func() {
		// we cannot use threading.GoSafe here, because of import cycle.
		ls.maybeCompressFile(file)
		ls.maybeDeleteOutdatedFiles()
	}()
}

func (ls *LoggerServer) rotate() error {
	if ls.fp != nil {
		err := ls.fp.Close()
		ls.fp = nil
		if err != nil {
			return err
		}
	}
	_, err := os.Stat(ls.filename)
	if err == nil && len(ls.backup) > 0 {
		backupFilename := ls.getBackupFilename()
		err = os.Rename(ls.filename, backupFilename)
		if err != nil {
			return err
		}
		ls.postRotate(backupFilename)
	}
	ls.backup = ls.rule.backupFilename()
	if ls.fp,err = os.Create(ls.filename);err != nil {
		closeOnExec(ls.fp)
	}
	return err
}

func (ls *LoggerServer) startWorker() {
	ls.waitGroup.Add(1)
	go func() {
		defer ls.waitGroup.Done()
		for {
			select {
			case event := <-ls.channel:
				ls.write(event)
			case <-ls.done:
				return
			}
		}
	}()
}

func (ls *LoggerServer) write(v []byte) {
	if ls.rule.shallRotate() {
		if err := ls.rotate(); err != nil {
			log.Println(err)
		} else {
			ls.rule.markRotated()
		}
	}
	if ls.fp != nil {
		if _, err := ls.fp.Write(v); err != nil {
			ErrorFormat("write ls.fp.Write err %+v", err)
		}
	}
}







