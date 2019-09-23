package mmh

import (
	"os"
	osexec "os/exec"

	"github.com/mritd/mmh/utils"
)

func Ping(tagOrName string) {

	server := findServerByName(tagOrName)
	if server == nil {
		utils.Exit("server not found", 1)
	}

	if server.Proxy == "" {
		cmd := osexec.Command("ping", server.Address)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		utils.CheckAndExit(cmd.Run())
	} else {
		Exec(tagOrName, "ping "+server.Address, true, true)
	}
}
