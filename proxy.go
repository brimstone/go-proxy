package proxy

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
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
	fmt.Printf("%s %s\n", r.Method, r.URL)
	req, err := http.NewRequest(r.Method, p.url.Scheme+"://"+p.url.Host+r.URL.Path, nil)
	fmt.Printf("r: %#v\n\n", req)
	resp, err := p.httpClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR Do: %s\n", err.Error())
		return
	}

	// [todo] - handle w.WriteHeader()
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ERROR ReadAll: %s\n", err.Error())
		return
	}
	w.Write(data)
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
