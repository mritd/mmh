package core

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// IgnoreProxy is set by the --no-proxy flag; when it is true the socks5 and
// http_proxy settings are ignored and the target is dialed directly
var IgnoreProxy bool

// dialSSH dials the ssh server address, if socks5 or http_proxy is configured
// on the server the connection is established through that proxy, otherwise
// mmh dials the address directly.
// Note: these settings only affect how the ssh server itself is reached, when
// "proxy"(jump server) is configured the jump server is dialed with its own
// settings, the target is reached through the jump's direct-tcpip channel.
func (s *Server) dialSSH() (net.Conn, error) {
	addr := net.JoinHostPort(s.Address, fmt.Sprint(s.Port))
	switch {
	// socks5 takes precedence when both are configured
	case s.Socks5 != "" && !IgnoreProxy:
		fmt.Fprintf(os.Stderr, "🧦 using socks5 proxy [%s] => %s(%s)\n", maskProxyAddr(s.Socks5), s.Name, addr)
		return socks5Dial(s.Socks5, addr, sshConnectTimeout)
	case s.HTTPProxy != "" && !IgnoreProxy:
		fmt.Fprintf(os.Stderr, "🌐 using http proxy [%s] => %s(%s)\n", maskProxyAddr(s.HTTPProxy), s.Name, addr)
		return httpProxyDial(s.HTTPProxy, addr, sshConnectTimeout)
	default:
		return net.DialTimeout("tcp", addr, sshConnectTimeout)
	}
}

// maskProxyAddr hides the credentials of a proxy address before printing it
func maskProxyAddr(addr string) string {
	if i := strings.LastIndex(addr, "@"); i >= 0 {
		return "***@" + addr[i+1:]
	}
	return addr
}

// socks5Dial dials target through the socks5 proxy; proxyAddr supports the
// "user:password@host:port" form
func socks5Dial(proxyAddr, target string, timeout time.Duration) (net.Conn, error) {
	var auth *proxy.Auth
	if i := strings.LastIndex(proxyAddr, "@"); i >= 0 {
		cred := strings.SplitN(proxyAddr[:i], ":", 2)
		auth = &proxy.Auth{User: cred[0]}
		if len(cred) == 2 {
			auth.Password = cred[1]
		}
		proxyAddr = proxyAddr[i+1:]
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("create socks5 dialer [%s] failed: %w", proxyAddr, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if cd, ok := dialer.(proxy.ContextDialer); ok {
		return cd.DialContext(ctx, "tcp", target)
	}
	return dialer.Dial("tcp", target)
}

// httpProxyDial dials target through an http proxy using the CONNECT method;
// proxyAddr supports both "host:port" and "http://user:password@host:port"
func httpProxyDial(proxyAddr, target string, timeout time.Duration) (net.Conn, error) {
	proxyURL, err := parseHTTPProxyURL(proxyAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", proxyURL.Host, timeout)
	if err != nil {
		return nil, fmt.Errorf("dial http proxy [%s] failed: %w", proxyURL.Host, err)
	}

	// the CONNECT handshake must not hang forever
	_ = conn.SetDeadline(time.Now().Add(timeout))

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: target},
		Host:   target,
		Header: make(http.Header),
	}
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		cred := base64.StdEncoding.EncodeToString([]byte(proxyURL.User.Username() + ":" + password))
		req.Header.Set("Proxy-Authorization", "Basic "+cred)
	}

	if err := req.Write(conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("write CONNECT request to http proxy failed: %w", err)
	}

	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("read CONNECT response from http proxy failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("http proxy CONNECT %s failed: %s", target, resp.Status)
	}
	_ = conn.SetDeadline(time.Time{})

	// data sent by the target right after the 200 response may already be
	// buffered by bufio, a plain conn would lose it
	if br.Buffered() > 0 {
		return &bufferedConn{Conn: conn, r: br}, nil
	}
	return conn, nil
}

// parseHTTPProxyURL parse the http proxy address, the scheme is optional
func parseHTTPProxyURL(proxyAddr string) (*url.URL, error) {
	if !strings.Contains(proxyAddr, "://") {
		proxyAddr = "http://" + proxyAddr
	}
	u, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid http proxy [%s]: %w", proxyAddr, err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid http proxy [%s]: host is empty", proxyAddr)
	}
	return u, nil
}

// bufferedConn is a net.Conn whose Read returns the already buffered bytes first
type bufferedConn struct {
	net.Conn
	r *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.r.Read(p)
}
