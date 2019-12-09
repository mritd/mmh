package mmh

import (
	"fmt"
	"io"
	"net"
)

// Tunnel will open an ssh tcp tunnel between the left address and the right address
func Tunnel(name, leftAddr, rightAddr string, reverse bool) {

	if !reverse {
		fmt.Printf("mmh tunnel listen at %s\n", leftAddr)
		listener, err := net.Listen("tcp", leftAddr)
		checkAndExit(err)
		defer func() { _ = listener.Close() }()

		for {
			leftConn, err := listener.Accept()
			if err != nil {
				fmt.Println("😱 " + err.Error())
				continue
			}

			fmt.Printf("new connection %s => [%s] => %s\n", leftConn.LocalAddr(), name, rightAddr)

			server, err := findServerByName(name)
			if err != nil {
				fmt.Println("😱 " + err.Error())
				continue
			}

			client, err := server.sshClient(false, true)
			if err != nil {
				fmt.Println("😱 " + err.Error())
				continue
			}

			rightConn, err := client.Dial("tcp", rightAddr)
			if err != nil {
				fmt.Println("😱 " + err.Error())
				continue
			}

			go connCopy(rightConn, leftConn)
		}
	} else {
		fmt.Printf("mmh reverse tunnel at [%s] %s\n", name, rightAddr)
		server, err := findServerByName(name)
		checkAndExit(err)
		client, err := server.sshClient(false, true)
		listener, err := client.Listen("tcp", rightAddr)
		checkAndExit(err)

		for {
			rightConn, err := listener.Accept()
			if err != nil {
				fmt.Println("😱 " + err.Error())
				continue
			}

			fmt.Printf("new connection %s:%s => [local] => %s\n", name, rightConn.RemoteAddr(), leftAddr)

			leftConn, err := net.Dial("tcp", leftAddr)
			if err != nil {
				fmt.Println("😱 " + err.Error())
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
		if err != nil {
			fmt.Println(err)
		}
	}()

	_, err := io.Copy(lc, rc)
	if err != nil {
		fmt.Println(err)
	}
}
