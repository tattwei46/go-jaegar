package main

import (
	"git.davidcheah.com/go-jaegar/opentracing"
	"io/ioutil"
	"log"
	"net/http"
	"fmt"
	"time"
)


type Server struct {
	Name string
	Port int
}

func (s *Server) GetAddress()string{
	return fmt.Sprintf("http://localhost:%d", s.Port)
}

var server1 = Server{
	Name: "server-1",
	Port: 8081,
}

var server2 = Server{
	Name: "server-2",
	Port: 8082,
}

func main() {
	opentracing.Init(server2.Name)
	s := &http.Server{
		Addr: fmt.Sprintf(":%d", server2.Port),
		Handler: opentracing.HttpMiddleware(server2.Name, http.HandlerFunc(handle)),
	}
	log.Fatal(s.ListenAndServe())
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("request received in server2")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	serverSpan,r:= opentracing.Deserialize(r,server2.Name)
	defer serverSpan.Finish()
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requester := string(reqBody)
	time.Sleep(500 * time.Millisecond)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello %s!", requester)
}