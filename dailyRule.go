package loggo

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type placeHolder struct{}

type loggerRule interface {
	backupFilename() string
	markRotated()
	outdatedFiles() []string
	shallRotate() bool
	getPrefix() string
}

type dailyRule struct {
	rotatedTime string
	filename    string
	prefix      string
	delimiter   string
	days        int
	gzip        bool
}


func defaultRule(filename, prefix, delimiter string, days int, gzip bool) loggerRule {
	return &dailyRule{
		rotatedTime: getNowDate(),
		filename:    filename,
		delimiter:   delimiter,
		days:        days,
		prefix:      prefix,
		gzip:        gzip,
	}
}

func (dr *dailyRule) backupFilename() string {
	return fmt.Sprintf("%s%s%s", dr.filename, dr.delimiter, getNowDate())
}
func (dr *dailyRule) markRotated() {
	dr.rotatedTime = getNowDate()
}
func (dr *dailyRule) outdatedFiles() []string {
	if dr.days <= 0 {
		return nil
	}
	var pattern string
	if dr.gzip {
		pattern = fmt.Sprintf("%s%s*.gz", dr.filename, dr.delimiter)
	} else {
		pattern = fmt.Sprintf("%s%s*", dr.filename, dr.delimiter)
	}
	files, err := filepath.Glob(pattern)
	if err != nil {
		ErrorFormat("failed to delete outdated log files, error: %s", err)
		return nil
	}
	var buf strings.Builder
	boundary := time.Now().Add(-time.Hour * time.Duration(HoursPerday*dr.days)).Format(DateFormat)
	if _, err := fmt.Fprintf(&buf, "%s%s%s", dr.filename, dr.delimiter, boundary); err != nil {
		ErrorFormat("failed to fmt.Fprintf %s%s%s err is %s", dr.filename, dr.delimiter, boundary, err)
		return nil
	}
	if dr.gzip {
		buf.WriteString(".gz")
	}
	boundaryFile := buf.String()
	var outDates []string
	for _, file := range files {
		if file < boundaryFile {
			outDates = append(outDates, file)
		}
	}
	return outDates
}
func (dr *dailyRule) shallRotate() bool {
	return len(dr.rotatedTime) > 0 && getNowDate() != dr.rotatedTime
}
func (dr *dailyRule) getPrefix() string {
	return dr.prefix
}

func getNowDate() string {
	return time.Now().Format(DateFormat)
}

func compressLogFile(file string) {
	start := time.Now()
	InfoFormat("compressing log file: %s", file)
	if err := gzipFile(file); err != nil {
		ErrorFormat("compress error: %s", err)
	} else {
		InfoFormat("compressed log file: %s, took %s", file, time.Since(start))
	}
}

func gzipFile(file string) error {
	in, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := in.Close(); err != nil {
			ErrorFormat("gzipFile.in.Close err %+v", err)
		}
	}()
	out, err := os.Create(fmt.Sprintf("%s.gz", file))
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			ErrorFormat("gzipFile.out.Close err %+v", err)
		}
	}()
	w := gzip.NewWriter(out)
	if _, err = io.Copy(w, in); err != nil {
		return err
	} else if err = w.Close(); err != nil {
		return err
	}
	return os.Remove(file)
}
