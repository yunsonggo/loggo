package loggo

type logOptions struct {
	gzipEnabled           bool
	logStackCoolDownMills int
	lastingDays           int
}

type logOption func(options *logOptions)

func handleOptions(opts []logOption) {
	for _, opt := range opts {
		opt(&options)
	}
}
