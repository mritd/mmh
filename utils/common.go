package utils

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

func CheckAndExit(err error) {
	if err != nil {
		fmt.Println("😱 " + err.Error())
		os.Exit(1)
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
	fmt.Println("😱 " + message)
	os.Exit(code)
}

func Root() bool {
	u, err := user.Current()
	CheckAndExit(err)
	return u.Uid == "0" || u.Gid == "0"
}
