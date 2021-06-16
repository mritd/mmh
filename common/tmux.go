package common

import (
	"fmt"
	"os"
	osexec "os/exec"
	"strings"
)

func Tmux() bool {
	return os.Getenv("TMUX") != ""
}

func TmuxSetWindowName(index, name string) {
	cmd := osexec.Command("tmux", "rename-window", "-t", index, name)
	PrintErr(cmd.Run())
}

func TmuxWindowInfo() (index, name string) {
	cmd := osexec.Command("tmux", "display-message", "-p", "#I #W")
	bs, err := cmd.CombinedOutput()
	if !CheckErr(err) {
		return "", ""
	}
	sp := strings.Fields(string(bs))
	if len(sp) != 2 {
		PrintErr(fmt.Errorf("failed to get tmux window info: %s", string(bs)))
		return "", ""
	}
	return sp[0], sp[1]
}

func TmuxSetAutomaticRename(index string, autoRename bool) {
	status := "on"
	if !autoRename {
		status = "off"
	}
	cmd := osexec.Command("tmux", "set-window", "-t", index, "automatic-rename", status)
	PrintErr(cmd.Run())
}

func TmuxAutomaticRename() bool {
	cmd := osexec.Command("tmux", "show-options", "-w")
	bs, err := cmd.CombinedOutput()
	if !CheckErr(err) {
		return false
	}
	if strings.Contains(string(bs), "automatic-rename on") {
		return true
	}
	if strings.Contains(string(bs), "automatic-rename off") {
		return false
	}

	cmd = osexec.Command("tmux", "show-options", "-gw")
	bs, err = cmd.CombinedOutput()
	if !CheckErr(err) {
		return false
	}
	if strings.Contains(string(bs), "automatic-rename on") {
		return true
	}

	return false
}
