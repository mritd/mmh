package core

import (
	"fmt"
	"io"
	"net"

	"github.com/mritd/mmh/common"
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
		client, err := server.wrapperClient(false)
		common.CheckAndExit(err)
		defer func() { _ = client.Close() }()

		for {
			leftConn, err := listener.Accept()
			if !common.CheckErr(err) {
				continue
			}

			fmt.Printf("new connection %s => [%s] => %s\n", leftConn.LocalAddr(), name, rightAddr)
			rightConn, err := client.Dial("tcp", rightAddr)
			if !common.CheckErr(err) {
				// client.Dial 失败时 rightConn 为 nil，已接受的 leftConn 必须关闭，
				// 否则每次失败都会泄漏一个本地 fd，最终导致 accept 报 "too many open files"
				_ = leftConn.Close()
				continue
			}

			go connCopy(rightConn, leftConn)
		}
	} else {
		fmt.Printf("mmh reverse tunnel listen at [%s] %s\n", name, rightAddr)
		server, err := findServerByName(name)
		common.CheckAndExit(err)
		client, err := server.wrapperClient(false)
		common.CheckAndExit(err)
		defer func() { _ = client.Close() }()
		listener, err := client.Listen("tcp", rightAddr)
		common.CheckAndExit(err)

		for {
			rightConn, err := listener.Accept()
			if err == io.EOF {
				fmt.Printf("Server [%s] ssh connection is closed, stop ssh tunnel.\n", server.Name)
				return
			}
			if !common.CheckErr(err) {
				continue
			}

			fmt.Printf("new connection %s:%s => [local] => %s\n", name, rightConn.RemoteAddr(), leftAddr)
			leftConn, err := net.Dial("tcp", leftAddr)
			if !common.CheckErr(err) {
				// 本地 dial 失败时 rightConn(远端 ssh channel) 必须关闭，
				// 否则每条失败请求都会在服务端留下一个已接受的连接
				_ = rightConn.Close()
				continue
			}
			go connCopy(leftConn, rightConn)
		}
	}

}

// closeWrite half-closes the write side of a connection if it supports it,
// so that EOF can be propagated to the peer (ssh.Channel and *net.TCPConn both do).
func closeWrite(c net.Conn) {
	if cw, ok := c.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
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
		// local client closed its write side, propagate EOF to the remote target,
		// otherwise the remote keep-alive connection (e.g. http proxy) never closes
		closeWrite(rc)
	}()

	_, err := io.Copy(lc, rc)
	common.PrintErr(err)
	closeWrite(lc)
}
