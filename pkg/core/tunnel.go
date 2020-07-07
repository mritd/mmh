package core

import (
	"fmt"
	"io"
	"net"

	"github.com/mritd/mmh/pkg/common"
)

// Tunnel will open an ssh tcp tunnel between the left address and the right address
func Tunnel(name, leftAddr, rightAddr string, reverse bool) {
	if !reverse {
		fmt.Printf("mmh tunnel listen at %s\n", leftAddr)
		listener, err := net.Listen("tcp", leftAddr)
		common.CheckAndExit(err)
		defer func() { _ = listener.Close() }()
		server, err := findServerByName(name)
		common.CheckAndExit(err)
		client, err := server.sshClient(false)
		common.CheckAndExit(err)

		for {
			leftConn, err := listener.Accept()
			if !common.CheckErr(err) {
				continue
			}

			fmt.Printf("new connection %s => [%s] => %s\n", leftConn.LocalAddr(), name, rightAddr)
			rightConn, err := client.Dial("tcp", rightAddr)
			if !common.CheckErr(err) {
				continue
			}

			go connCopy(rightConn, leftConn)
		}
	} else {
		fmt.Printf("mmh reverse tunnel at [%s] %s\n", name, rightAddr)
		server, err := findServerByName(name)
		common.CheckAndExit(err)
		client, err := server.sshClient(false)
		common.CheckAndExit(err)
		listener, err := client.Listen("tcp", rightAddr)
		common.CheckAndExit(err)

		for {
			rightConn, err := listener.Accept()
			if !common.CheckErr(err) {
				continue
			}

			fmt.Printf("new connection %s:%s => [local] => %s\n", name, rightConn.RemoteAddr(), leftAddr)
			leftConn, err := net.Dial("tcp", leftAddr)
			if !common.CheckErr(err) {
				continue
			}
			go connCopy(leftConn, rightConn)
		}
	}

}

func connCopy(rc, lc net.Conn) {
	defer func() {
		_ = rc.Close()
		_ = lc.Close()
	}()

	go func() {
		_, err := io.Copy(rc, lc)
		common.PrintErr(err)
	}()

	_, err := io.Copy(lc, rc)
	common.PrintErr(err)
}
