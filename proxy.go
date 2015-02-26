package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Proxy struct {
	url        *url.URL
	httpClient *http.Client
}

func (p *Proxy) Handle(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, p.url.Scheme+"://"+p.url.Host+r.URL.String(), nil)
	resp, err := p.httpClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR Do: %s\n", err.Error())
		return
	}

	w.Header().Set("Content-Type", resp.Header["Content-Type"][0])

	defer resp.Body.Close()

	// If this is totally normal content
	if resp.ContentLength != -1 {
		io.Copy(w, resp.Body)
		return
	}

	// If this is chunked, handle it that way
	buf := make([]byte, 65535)
	size := 0
	for {
		size, err = resp.Body.Read(buf)
		w.Write(buf[0:size])
		w.(http.Flusher).Flush()
		if size < len(buf) {
			break
		}
	}
}

func newHTTPClient(u *url.URL, tlsConfig *tls.Config, timeout time.Duration) *http.Client {
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	switch u.Scheme {
	default:
		httpTransport.Dial = func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		}
	case "unix":
		socketPath := u.Path
		unixDial := func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}
		httpTransport.Dial = unixDial
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	}
	return &http.Client{Transport: httpTransport}
}

func New(u string) (*Proxy, error) {

	proxyClient := new(Proxy)

	// [todo] - add error handling here
	uri, _ := url.Parse(u)

	proxyClient.httpClient = newHTTPClient(uri, nil, time.Second)
	proxyClient.url = uri

	return proxyClient, nil
}
