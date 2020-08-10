package common

import (
	"os"
	osexec "os/exec"
	"os/user"
	"strings"
	"text/template"

	"github.com/mritd/promptx"

	"fmt"
)

func init() {
	promptx.FuncMap["maxLen"] = maxLen
	promptx.FuncMap["mergeTags"] = mergeTags
}

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
		fmt.Println(prefix + ": 😱 " + err.Error())
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

func maxLen(length int, name string) string {
	if length == 0 {
		length = 15
	}
	sTpl := fmt.Sprintf("%%-%ds", length)
	if len(name) < length {
		return fmt.Sprintf(sTpl, name)
	} else {
		return fmt.Sprintf(sTpl, shortenString(name, length))
	}
}

func shortenString(str string, n int) string {
	if len(str) <= n {
		return str
	} else {
		return str[:n]
	}
}

// merge tags
func mergeTags(tags []string) string {
	return strings.Join(tags, ",")
}

func Template(tpl string) (*template.Template, error) {
	return template.New("").Funcs(promptx.FuncMap).Parse(tpl)
}

func CMD(cmd string) (string, []string) {
	cmds := strings.Fields(cmd)
	if len(cmds) > 1 {
		return cmds[0], cmds[1:]
	} else {
		return cmds[0], nil
	}
}

func Tmux() bool {
	return os.Getenv("TMUX") != ""
}

func TmuxSetWindowName(name string) {
	cmd := osexec.Command("tmux", "rename-window", name)
	CheckErr(cmd.Run())
}

func TmuxWindowName() string {
	cmd := osexec.Command("tmux", "display-message", "-p", "#W")
	bs, err := cmd.CombinedOutput()
	if CheckErr(err) {
		return ""
	}
	return strings.TrimSpace(string(bs))
}

func TmuxSetAutomaticRename(autoRename bool) {
	status := "on"
	if !autoRename {
		status = "off"
	}
	cmd := osexec.Command("tmux", "set-window", "automatic-rename", status)
	CheckErr(cmd.Run())
}

func TmuxAutomaticRename() bool {
	cmd := osexec.Command("tmux", "show-options", "-gw")
	bs, err := cmd.CombinedOutput()
	if CheckErr(err) {
		return false
	}
	return strings.Contains(string(bs), "automatic-rename on")
}
