package proxy

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"regexp"
	"strings"
)

type handler struct {
	path *regexp.Regexp
	dst  *url.URL
	close bool
}
type Proxy struct {
	listen   string
	listener net.Listener
	handlers []handler
}

func (p *Proxy) ListenAndServe(l string) {
	// start listening on our interface and port
	var err error
	p.listener, err = net.Listen("tcp", l)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		go p.handleConnection(conn)
	}
}

func (p *Proxy) handleConnection(c net.Conn) {
	defer c.Close()

	var outbound net.Conn
	var err error

	header, err := readUntil(c, "\n")
	if err != nil {
		fmt.Println("Error reading first line from connection:", err.Error())
		return
	}

	// parse out method, path, and protocol
	requestArray := strings.Split(string(header), " ")
	for h := 0; h < len(p.handlers); h++ {
		if p.handlers[h].path.MatchString(requestArray[1]) {
			if p.handlers[h].dst.Scheme == "http" {
				outbound, err = net.Dial("tcp", p.handlers[h].dst.Host)
			} else if p.handlers[h].dst.Scheme == "unix" {
				outbound, err = net.Dial("unix", p.handlers[h].dst.Path)
			} else {
				fmt.Println("go-proxy can't handle", p.handlers[h].dst.Scheme)
				return
			}
			if err != nil {
				fmt.Println("Error opening connection to", p.handlers[h].dst, err.Error())
				return
			}
			fmt.Fprintf(outbound, "%s", string(header))
			if p.handlers[h].close {
				fmt.Fprintf(outbound, "Connection: Close\n")
			}
			go io.Copy(outbound, c)
			io.Copy(c, outbound)
			return
		}
	}

	return
}

func readUntil(r net.Conn, delimiter string) (string, error) {
	var output []byte
	for {
		buffer := make([]byte, 1)
		_, err := r.Read(buffer)
		if err != nil {
			fmt.Println("Error while copying from socket", err.Error())
			return "", err
		}
		output = append(output, buffer[0])
		if string(buffer) == delimiter {
			return string(output), nil
		}
	}

}

func (p *Proxy) Handle(path string, d string, c bool) error {
	// [todo] - push this onto p.handlers
	uri, err := url.Parse(d)
	if err == nil {
		newHandler := handler{path: regexp.MustCompile(path), dst: uri, close: c}
		p.handlers = append(p.handlers, newHandler)
	}
	return err
}

func New() (*Proxy, error) {

	proxyClient := new(Proxy)

	return proxyClient, nil
}
