package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	proxy, err := New()
	if err != nil {
		t.Errorf("New() ERROR Creating proxy object %q", err.Error())
	}

	if proxy == nil {
		t.Errorf("New() ERROR proxy object not defined.")
	}

	proxy.Handle("/", "unix:///var/run/docker.sock")
}

// Back server
type backServe struct{}

func (self backServe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

// except server
type exceptServe struct{}

func (self exceptServe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "proxy")
}

func TestRequst(t *testing.T) {

	// setup our "Real" server
	back := &http.Server{Addr: "127.0.0.1:9000", Handler: new(backServe)}
	go back.ListenAndServe()

	except := &http.Server{Addr: "127.0.0.1:9001", Handler: new(exceptServe)}
	go except.ListenAndServe()

	proxy, err := New()
	if err != nil {
		t.Errorf("Get(): %q", err.Error())
		return
	}

	fmt.Println("Adding route for /proxy to :9001")
	err = proxy.Handle("^/proxy", "http://localhost:9001")
	if err != nil {
		t.Errorf("Get(): %q", err.Error())
		return
	}

	fmt.Println("Adding route for / to :9000")
	err = proxy.Handle("^/$", "http://localhost:9000")
	if err != nil {
		t.Errorf("Get(): %q", err.Error())
		return
	}

	fmt.Println("Starting proxy at :9002")
	go proxy.ListenAndServe(":9002")

	fmt.Println("Requesting http://127.0.0.1:9002/")
	result, err := http.Get("http://127.0.0.1:9002/")
	if err != nil {
		t.Errorf("Get(): %q", err.Error())
		return
	}
	defer result.Body.Close()
	fmt.Println("Reading all of body")
	data, err := ioutil.ReadAll(result.Body)
	if string(data) != "ok" {
		t.Errorf("Result not \"ok\"", string(data))
		return
	}

	fmt.Println("Requesting http://127.0.0.1:9002/proxy")
	result, err = http.Get("http://127.0.0.1:9002/proxy")
	if err != nil {
		t.Errorf("Get(): %q", err.Error())
		return
	}
	defer result.Body.Close()
	data, err = ioutil.ReadAll(result.Body)
	if string(data) != "proxy" {
		t.Errorf("Result not \"proxy\"", string(data))
		return
	}

}
