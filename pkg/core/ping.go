package core

import (
	"os"
	osexec "os/exec"

	"github.com/mritd/mmh/pkg/common"
)

// Ping execute the ping target server
// if the target server requires a proxy, ping on the last proxy server
func Ping(name string) {
	server, err := findServerByName(name)
	common.CheckAndExit(err)

	if server.Proxy == "" {
		cmd := osexec.Command("ping", server.Address)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		common.CheckAndExit(cmd.Run())
	} else {
		Exec("ping "+server.Address, name, false, true)
	}
}
