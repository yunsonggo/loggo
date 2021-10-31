package loggo

import "testing"

func TestLoggo(t *testing.T) {
	Init(Config{Path: "logs",Stdout: true})
	Slow("Slow")
	Stat("a")
	Info("info test")
	Error("error test")
	res := map[int]string{1:"1111",2:"2222"}
	InfoFormat("InfoFormat %s,%+v","info format",res)
	ErrorFormat("ErrorFormat: %s,%+v","error format",res)
	SlowFormat("SlowFormat: %s,%+v","slow format",res)
	StatFormat("StatFormat: %s,%+v","stat format",res)
}