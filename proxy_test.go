package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	proxy, err := New("unix:///var/run/docker.sock")
	if err != nil {
		t.Errorf("New() ERROR Creating proxy object %q", err.Error())
	}

	if proxy == nil {
		t.Errorf("New() ERROR proxy object not defined.")
	}
}

// Back server
type backServe struct{}

func (self backServe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

// Front server
type frontServe struct{}

func (self frontServe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy, err := New("http://localhost:9000")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	proxy.Handle(w, r)
}

func TestRequst(t *testing.T) {

	// setup our "Real" server
	back := &http.Server{Addr: "127.0.0.1:9000", Handler: new(backServe)}
	go back.ListenAndServe()

	front := &http.Server{Addr: "127.0.0.1:9001", Handler: new(frontServe)}
	go front.ListenAndServe()

	result, err := http.Get("http://127.0.0.1:9001/")
	if err != nil {
		t.Errorf("Get(): %q", err.Error())
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if string(data) != "ok" {
		t.Errorf("Result not \"ok\"")
	}

}
