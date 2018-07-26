package utils

import "log"

func CheckAndExit(err error) {
	if err != nil {
		log.Panic(err)
	}
}
