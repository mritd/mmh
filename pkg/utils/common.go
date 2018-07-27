package utils

import (
	"fmt"
	"os"
	"strings"
)

func CheckAndExit(err error) {
	if err != nil {
		panic(err)
	}
}

func ShortenString(str string, n int) string {
	if len(str) <= n {
		return str
	} else {
		return str[:n]
	}
}

func Exit(message string, code int) {
	if strings.TrimSpace(message) == "" {
		message = "No message"
	}
	fmt.Println(message)
	os.Exit(code)
}
