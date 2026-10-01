package webpool

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
)

func TestDialRawConnConnectAuthAndBufferedBytes(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	serverErr := make(chan error, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverErr <- acceptErr
			return
		}
		defer conn.Close()
		request, readErr := http.ReadRequest(bufio.NewReader(conn))
		if readErr != nil {
			serverErr <- readErr
			return
		}
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:secret"))
		if request.Method != http.MethodConnect || request.Host != "example.com:443" || request.Header.Get("Proxy-Authorization") != wantAuth {
			serverErr <- fmt.Errorf("unexpected CONNECT request: method=%s host=%s auth=%q", request.Method, request.Host, request.Header.Get("Proxy-Authorization"))
			return
		}
		_, writeErr := fmt.Fprint(conn, "HTTP/1.1 200 Connection Established\r\n\r\nhello")
		serverErr <- writeErr
	}()

	proxyURL, err := url.Parse("http://user:secret@" + listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn, err := dialRawConn(context.Background(), &net.Dialer{}, proxyURL, "tcp", "example.com:443")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	buf := make([]byte, 5)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "hello" {
		t.Fatalf("tunnel data = %q, want hello", buf)
	}
	if err := <-serverErr; err != nil {
		t.Fatal(err)
	}
}

func TestProxyAddressHandlesDefaultsAndIPv6(t *testing.T) {
	proxy, err := url.Parse("http://[::1]")
	if err != nil {
		t.Fatal(err)
	}
	got, err := proxyAddress(proxy, "http")
	if err != nil {
		t.Fatal(err)
	}
	if got != "[::1]:80" {
		t.Fatalf("proxy address = %q, want [::1]:80", got)
	}
}
