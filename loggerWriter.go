package loggo

import "log"

type logWriter struct {
	logger *log.Logger
}

func newLogWriter(lg *log.Logger) logWriter {
	return logWriter{
		logger: lg,
	}
}

func (lw *logWriter) Write(data []byte) (int, error) {
	lw.logger.Print(string(data))
	return len(data), nil
}

func (lw *logWriter) Close() error {
	return nil
}
