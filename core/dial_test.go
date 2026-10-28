package core

import (
	"bufio"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

const dialTestTimeout = 3 * time.Second

// fakeProxyHost can't be resolved, so it is only reachable through the fake
// proxies below. If a connection ever bypasses the proxy, the tests using it
// fail instead of silently dialing the local target directly.
const fakeProxyHost = "mmh.test.invalid"

// testProxy records the targets requested from a fake proxy server, so a test
// can assert the proxy was really used
type testProxy struct {
	addr string

	mu      sync.Mutex
	targets []string
}

func (p *testProxy) record(target string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.targets = append(p.targets, target)
}

func (p *testProxy) requests() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.targets...)
}

func assertProxyHit(t *testing.T, p *testProxy, want string) {
	t.Helper()

	reqs := p.requests()
	if len(reqs) != 1 {
		t.Fatalf("expected exactly 1 connection through the proxy, got %d: %v", len(reqs), reqs)
	}
	if reqs[0] != want {
		t.Fatalf("proxy was asked for %q, want %q", reqs[0], want)
	}
}

func parseAddr(t *testing.T, addr string) (string, int) {
	t.Helper()

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host port failed: %v", err)
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("parse port failed: %v", err)
	}
	return host, p
}

// startEchoServer starts a tcp server which echoes everything back
func startEchoServer(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen echo server failed: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				_, _ = io.Copy(c, c)
			}(conn)
		}
	}()
	return ln.Addr().String()
}

// startSOCKS5Server starts a minimal socks5 server; user/password enables the
// username/password sub negotiation (RFC 1929). mapping redirects the host
// requested by the client to the address the proxy really dials, which lets
// tests point fakeProxyHost at a local listener.
func startSOCKS5Server(t *testing.T, user, password string, mapping map[string]string) *testProxy {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen socks5 server failed: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	p := &testProxy{addr: ln.Addr().String()}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				_ = serveSOCKS5(c, p, user, password, mapping)
			}(conn)
		}
	}()
	return p
}

func serveSOCKS5(c net.Conn, p *testProxy, user, password string, mapping map[string]string) error {
	br := bufio.NewReader(c)

	// method negotiation
	head := make([]byte, 2)
	if _, err := io.ReadFull(br, head); err != nil {
		return err
	}
	if head[0] != 0x05 {
		return fmt.Errorf("unexpected socks version: %d", head[0])
	}
	methods := make([]byte, int(head[1]))
	if _, err := io.ReadFull(br, methods); err != nil {
		return err
	}

	want := byte(0x00)
	if user != "" {
		want = 0x02
	}
	acceptable := false
	for _, m := range methods {
		if m == want {
			acceptable = true
		}
	}
	if !acceptable {
		_, _ = c.Write([]byte{0x05, 0xff})
		return fmt.Errorf("no acceptable auth method")
	}
	if _, err := c.Write([]byte{0x05, want}); err != nil {
		return err
	}

	// username/password sub negotiation
	if want == 0x02 {
		ver := make([]byte, 2)
		if _, err := io.ReadFull(br, ver); err != nil {
			return err
		}
		uname := make([]byte, int(ver[1]))
		if _, err := io.ReadFull(br, uname); err != nil {
			return err
		}
		plen := make([]byte, 1)
		if _, err := io.ReadFull(br, plen); err != nil {
			return err
		}
		pass := make([]byte, int(plen[0]))
		if _, err := io.ReadFull(br, pass); err != nil {
			return err
		}
		if string(uname) != user || string(pass) != password {
			_, _ = c.Write([]byte{0x01, 0x01})
			return fmt.Errorf("bad credentials")
		}
		if _, err := c.Write([]byte{0x01, 0x00}); err != nil {
			return err
		}
	}

	// connect request
	req := make([]byte, 4)
	if _, err := io.ReadFull(br, req); err != nil {
		return err
	}
	if req[1] != 0x01 {
		return fmt.Errorf("unsupported command: %d", req[1])
	}

	var host string
	switch req[3] {
	case 0x01:
		ip := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(br, ip); err != nil {
			return err
		}
		host = net.IP(ip).String()
	case 0x03:
		l := make([]byte, 1)
		if _, err := io.ReadFull(br, l); err != nil {
			return err
		}
		name := make([]byte, int(l[0]))
		if _, err := io.ReadFull(br, name); err != nil {
			return err
		}
		host = string(name)
	case 0x04:
		ip := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(br, ip); err != nil {
			return err
		}
		host = net.IP(ip).String()
	default:
		return fmt.Errorf("unsupported address type: %d", req[3])
	}
	port := make([]byte, 2)
	if _, err := io.ReadFull(br, port); err != nil {
		return err
	}
	target := net.JoinHostPort(host, strconv.Itoa(int(binary.BigEndian.Uint16(port))))
	p.record(target)
	if mapped, ok := mapping[host]; ok {
		target = mapped
	}

	remote, err := net.DialTimeout("tcp", target, dialTestTimeout)
	if err != nil {
		_, _ = c.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return err
	}
	defer func() { _ = remote.Close() }()
	if _, err := c.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}

	// bytes buffered while reading the request belong to the tunnel
	if n := br.Buffered(); n > 0 {
		buffered := make([]byte, n)
		if _, err := io.ReadFull(br, buffered); err != nil {
			return err
		}
		if _, err := remote.Write(buffered); err != nil {
			return err
		}
	}

	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(remote, br)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(c, remote)
		done <- struct{}{}
	}()
	<-done
	return nil
}

// startHTTPProxyServer starts a minimal http proxy which only supports CONNECT
func startHTTPProxyServer(t *testing.T, user, password string, mapping map[string]string) *testProxy {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen http proxy server failed: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	p := &testProxy{addr: ln.Addr().String()}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				_ = serveHTTPConnect(c, p, user, password, mapping)
			}(conn)
		}
	}()
	return p
}

func serveHTTPConnect(c net.Conn, p *testProxy, user, password string, mapping map[string]string) error {
	br := bufio.NewReader(c)
	req, err := http.ReadRequest(br)
	if err != nil {
		return err
	}
	if req.Method != http.MethodConnect {
		_, _ = io.WriteString(c, "HTTP/1.1 405 Method Not Allowed\r\n\r\n")
		return fmt.Errorf("unsupported method: %s", req.Method)
	}
	if user != "" {
		want := "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
		if req.Header.Get("Proxy-Authorization") != want {
			_, _ = io.WriteString(c, "HTTP/1.1 407 Proxy Authentication Required\r\n\r\n")
			return fmt.Errorf("bad proxy credentials")
		}
	}

	p.record(req.Host)
	target := req.Host
	if host, _, err := net.SplitHostPort(req.Host); err == nil {
		if mapped, ok := mapping[host]; ok {
			target = mapped
		}
	}

	remote, err := net.DialTimeout("tcp", target, dialTestTimeout)
	if err != nil {
		_, _ = io.WriteString(c, "HTTP/1.1 502 Bad Gateway\r\n\r\n")
		return err
	}
	defer func() { _ = remote.Close() }()

	if _, err := io.WriteString(c, "HTTP/1.1 200 Connection established\r\n\r\n"); err != nil {
		return err
	}
	if n := br.Buffered(); n > 0 {
		buffered := make([]byte, n)
		if _, err := io.ReadFull(br, buffered); err != nil {
			return err
		}
		if _, err := remote.Write(buffered); err != nil {
			return err
		}
	}

	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(remote, br)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(c, remote)
		done <- struct{}{}
	}()
	<-done
	return nil
}

func assertEcho(t *testing.T, conn net.Conn) {
	t.Helper()
	defer func() { _ = conn.Close() }()

	if err := conn.SetDeadline(time.Now().Add(dialTestTimeout)); err != nil {
		t.Fatalf("set deadline failed: %v", err)
	}
	msg := []byte("hello mmh proxy")
	if _, err := conn.Write(msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	buf := make([]byte, len(msg))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(buf) != string(msg) {
		t.Fatalf("echo mismatch: got %q, want %q", buf, msg)
	}
}

func TestSocks5Dial(t *testing.T) {
	echoAddr := startEchoServer(t)
	p := startSOCKS5Server(t, "", "", nil)

	conn, err := socks5Dial(p.addr, echoAddr, dialTestTimeout)
	if err != nil {
		t.Fatalf("socks5 dial failed: %v", err)
	}
	assertEcho(t, conn)
	assertProxyHit(t, p, echoAddr)
}

func TestSocks5DialWithAuth(t *testing.T) {
	echoAddr := startEchoServer(t)
	p := startSOCKS5Server(t, "mmh", "secret", nil)

	conn, err := socks5Dial("mmh:secret@"+p.addr, echoAddr, dialTestTimeout)
	if err != nil {
		t.Fatalf("socks5 dial with auth failed: %v", err)
	}
	assertEcho(t, conn)
	assertProxyHit(t, p, echoAddr)
}

func TestSocks5DialWithWrongAuth(t *testing.T) {
	echoAddr := startEchoServer(t)
	p := startSOCKS5Server(t, "mmh", "secret", nil)

	if conn, err := socks5Dial("mmh:wrong@"+p.addr, echoAddr, dialTestTimeout); err == nil {
		_ = conn.Close()
		t.Fatal("expected auth failure, got a working connection")
	}
}

func TestHTTPProxyDial(t *testing.T) {
	echoAddr := startEchoServer(t)
	p := startHTTPProxyServer(t, "", "", nil)

	conn, err := httpProxyDial(p.addr, echoAddr, dialTestTimeout)
	if err != nil {
		t.Fatalf("http proxy dial failed: %v", err)
	}
	assertEcho(t, conn)
	assertProxyHit(t, p, echoAddr)
}

func TestHTTPProxyDialWithAuth(t *testing.T) {
	echoAddr := startEchoServer(t)
	p := startHTTPProxyServer(t, "mmh", "secret", nil)

	conn, err := httpProxyDial("http://mmh:secret@"+p.addr, echoAddr, dialTestTimeout)
	if err != nil {
		t.Fatalf("http proxy dial with auth failed: %v", err)
	}
	assertEcho(t, conn)
	assertProxyHit(t, p, echoAddr)
}

// startSSHServer starts a minimal ssh server which echoes the executed command back
func startSSHServer(t *testing.T, user, password string) string {
	t.Helper()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate host key failed: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("create host signer failed: %v", err)
	}

	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == user && string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid credentials")
		},
	}
	cfg.AddHostKey(signer)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen ssh server failed: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				_, chans, reqs, err := ssh.NewServerConn(c, cfg)
				if err != nil {
					return
				}
				go ssh.DiscardRequests(reqs)
				for newChan := range chans {
					if newChan.ChannelType() != "session" {
						_ = newChan.Reject(ssh.UnknownChannelType, "only session is supported")
						continue
					}
					ch, requests, err := newChan.Accept()
					if err != nil {
						return
					}
					go func() {
						defer func() { _ = ch.Close() }()
						for req := range requests {
							if req.Type != "exec" {
								_ = req.Reply(false, nil)
								continue
							}
							var payload struct{ Command string }
							_ = ssh.Unmarshal(req.Payload, &payload)
							_ = req.Reply(true, nil)
							_, _ = io.WriteString(ch, "echo: "+payload.Command)
							_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
							return
						}
					}()
				}
			}(conn)
		}
	}()
	return ln.Addr().String()
}

// assertSSHExec makes sure the ssh client returned by Server.ssh can run a command
func assertSSHExec(t *testing.T, s *Server) {
	t.Helper()

	client, err := s.ssh(false, 0)
	if err != nil {
		t.Fatalf("ssh connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	out, err := session.Output("hello")
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if string(out) != "echo: hello" {
		t.Fatalf("unexpected exec output: %q", out)
	}
}

// TestSSHDirectCannotReachFakeHost proves fakeProxyHost is unreachable without
// a proxy, so the proxy tests can't pass by dialing the target directly
func TestSSHDirectCannotReachFakeHost(t *testing.T) {
	s := &Server{Address: fakeProxyHost, Port: 22, User: "mmh", Password: "secret"}

	client, err := s.ssh(false, 0)
	if err == nil {
		_ = client.Close()
		t.Fatalf("expected dialing %s directly to fail", fakeProxyHost)
	}
}

func TestSSHConnectionDirect(t *testing.T) {
	sshAddr := startSSHServer(t, "mmh", "secret")
	host, port := parseAddr(t, sshAddr)

	assertSSHExec(t, &Server{Address: host, Port: port, User: "mmh", Password: "secret"})
}

func TestSSHConnectionThroughSocks5(t *testing.T) {
	sshAddr := startSSHServer(t, "mmh", "secret")
	_, port := parseAddr(t, sshAddr)

	p := startSOCKS5Server(t, "", "", map[string]string{fakeProxyHost: sshAddr})
	s := &Server{Address: fakeProxyHost, Port: port, User: "mmh", Password: "secret", Socks5: p.addr}
	assertSSHExec(t, s)
	assertProxyHit(t, p, net.JoinHostPort(fakeProxyHost, strconv.Itoa(port)))
}

func TestSSHConnectionThroughHTTPProxy(t *testing.T) {
	sshAddr := startSSHServer(t, "mmh", "secret")
	_, port := parseAddr(t, sshAddr)

	p := startHTTPProxyServer(t, "", "", map[string]string{fakeProxyHost: sshAddr})
	s := &Server{Address: fakeProxyHost, Port: port, User: "mmh", Password: "secret", HTTPProxy: p.addr}
	assertSSHExec(t, s)
	assertProxyHit(t, p, net.JoinHostPort(fakeProxyHost, strconv.Itoa(port)))
}

// TestDialSSHPrefersSocks5 makes sure socks5 wins when both proxies are
// configured: the target is only reachable through socks5, the http proxy
// points to a closed port and would fail if it were used
func TestDialSSHPrefersSocks5(t *testing.T) {
	echoAddr := startEchoServer(t)
	_, port := parseAddr(t, echoAddr)
	socks := startSOCKS5Server(t, "", "", map[string]string{fakeProxyHost: echoAddr})

	dead, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	deadAddr := dead.Addr().String()
	_ = dead.Close()

	s := &Server{Address: fakeProxyHost, Port: port, Socks5: socks.addr, HTTPProxy: deadAddr}
	conn, err := s.dialSSH()
	if err != nil {
		t.Fatalf("dial ssh failed: %v", err)
	}
	assertEcho(t, conn)
	assertProxyHit(t, socks, net.JoinHostPort(fakeProxyHost, strconv.Itoa(port)))
}

func TestMaskProxyAddr(t *testing.T) {
	tests := map[string]string{
		"127.0.0.1:1080":                    "127.0.0.1:1080",
		"user:secret@127.0.0.1:1080":        "***@127.0.0.1:1080",
		"http://user:secret@127.0.0.1:8080": "***@127.0.0.1:8080",
	}
	for addr, want := range tests {
		if got := maskProxyAddr(addr); got != want {
			t.Errorf("maskProxyAddr(%q) = %q, want %q", addr, got, want)
		}
	}
}

// TestDialSSHIgnoresProxy makes sure the --no-proxy flag bypasses the configured
// proxy: the dead proxy is used by default, and the target is dialed directly
// when IgnoreProxy is set
func TestDialSSHIgnoresProxy(t *testing.T) {
	echoAddr := startEchoServer(t)
	host, port := parseAddr(t, echoAddr)

	dead, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	deadAddr := dead.Addr().String()
	_ = dead.Close()

	s := &Server{Address: host, Port: port, Socks5: deadAddr}

	old := IgnoreProxy
	defer func() { IgnoreProxy = old }()

	IgnoreProxy = false
	if conn, err := s.dialSSH(); err == nil {
		_ = conn.Close()
		t.Fatal("expected the dead socks5 proxy to be used and fail")
	}

	IgnoreProxy = true
	conn, err := s.dialSSH()
	if err != nil {
		t.Fatalf("dial ssh with --no-proxy failed: %v", err)
	}
	assertEcho(t, conn)
}
