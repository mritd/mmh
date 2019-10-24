package mmh

import (
	"fmt"
	"io"
	"net"

	"github.com/mritd/mmh/utils"
)

// Tunnel will open an ssh tcp tunnel between the local port and the remote port
func Tunnel(name, localAddr, remoteAddr string) {

	fmt.Printf("mmh tunnel listen at %s\n", localAddr)
	listener, err := net.Listen("tcp", localAddr)
	utils.CheckAndExit(err)
	defer func() { _ = listener.Close() }()

	for {
		localConn, err := listener.Accept()
		utils.CheckAndExit(err)

		fmt.Printf("new connection %s => [%s] => %s\n", localConn.LocalAddr(), name, remoteAddr)

		server, err := findServerByName(name)
		utils.CheckAndExit(err)
		client, err := server.sshClient(false, true)
		utils.CheckAndExit(err)
		remoteConn, err := client.Dial("tcp", remoteAddr)
		utils.CheckAndExit(err)

		go func() {
			_, err := io.Copy(remoteConn, localConn)
			if err != nil {
				fmt.Println(err)
			}
		}()

		go func() {
			_, err := io.Copy(localConn, remoteConn)
			if err != nil {
				fmt.Println(err)
			}
		}()

	}

}
