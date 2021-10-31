package loggo

import "log"

func Init(c Config) {
	if err := loadConfig(c); err != nil {
		log.Fatal(err)
	}
}
