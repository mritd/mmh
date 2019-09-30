package mmh

import (
	"os"
	osexec "os/exec"

	"github.com/mritd/mmh/utils"
)

// Ping execute the ping target server
// if the target server requires a proxy, ping on the last proxy server
func Ping(name string) {

	server, err := findServerByName(name)
	utils.CheckAndExit(err)

	if server.Proxy == "" {
		cmd := osexec.Command("ping", server.Address)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		utils.CheckAndExit(cmd.Run())
	} else {
		Exec(name, "ping "+server.Address, true, true)
	}
}
