package mmh

import (
	"os"
	osexec "os/exec"
)

// Ping execute the ping target server
// if the target server requires a proxy, ping on the last proxy server
func Ping(name string) {

	server, err := findServerByName(name)
	checkAndExit(err)

	if server.Proxy == "" {
		cmd := osexec.Command("ping", server.Address)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		checkAndExit(cmd.Run())
	} else {
		Exec(name, "ping "+server.Address, true, true)
	}
}
