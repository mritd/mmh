package common

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

func CheckAndExit(err error) {
	PrintErr(err)
	if err != nil {
		os.Exit(1)
	}
}

func CheckErr(err error) bool {
	PrintErr(err)
	return err == nil
}

func PrintErrWithPrefix(prefix string, err error) {
	if err != nil {
		fmt.Println(prefix, err.Error())
	}
}

func PrintErr(err error) {
	if err != nil {
		fmt.Println("😱 " + err.Error())
	}
}

func IsRoot() bool {
	u, err := user.Current()
	CheckAndExit(err)
	return u.Uid == "0" || u.Gid == "0"
}

func Exit(message string, code int) {
	if strings.TrimSpace(message) == "" {
		message = "No message"
	}
	fmt.Println("😱 " + message)
	os.Exit(code)
}

func ParseCommand(cmd string) (string, []string) {
	cs := strings.Fields(cmd)
	if len(cs) > 1 {
		return cs[0], cs[1:]
	} else {
		return cs[0], nil
	}
}
