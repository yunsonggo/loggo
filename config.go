package loggo

const (
	TimeFormat       = "2006-01-02 15:04:05.000" // 日期格式
	AccessFile       = "access.log"
	ErrorFile        = "error.log"
	SlowFile         = "slow.log" // 超时日志
	StateFile        = "state.log"
	VARMODE          = "var"
	DefaultHostname  = "loggo"
	InfoPrefix       = "[I]"
	ErrorPrefix      = "[E]"
	SlowPrefix       = "[T]"
	StackPrefix      = "[P]"
	StatePrefix      = "[S]"
	FileDelimiter    = "-"
	CallerInnerdepth = 5
)

type Config struct {
	NameSpace           string `json:",optional"`
	Stdout              bool   `json:"stdout,default=true"`
	LogMode             string `json:",options=regular|volume,default=regular"`
	Path                string `json:",default=logs"`
	Compress            bool   `json:",optional"`
	LastingDays         int    `json:",optional"`
	StackCoolDownMillis int    `json:",default=100"`
}
