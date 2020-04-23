package main

import (
	"context"
	"fmt"
	"git.davidcheah.com/go-jaegar/opentracing"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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

func main(){
	opentracing.Init(server1.Name)
	s := &http.Server{
		Addr: fmt.Sprintf(":%d", server1.Port),
		Handler: opentracing.HttpMiddleware(server1.Name, http.HandlerFunc(handle)),
	}
	log.Fatal(s.ListenAndServe())
}

func handle(w http.ResponseWriter, r *http.Request){
	fmt.Println("request received in server1")
	t := time.Now()
	requester := r.URL.Path[1:]
	time.Sleep(250*time.Millisecond)
	makeRequest(r.Context(), server2.GetAddress(), requester)
	fmt.Fprintf(w, "Hello %s! Time is %s", requester, t.Format("Mon Jan _2 15:04:05 2006"))
}

func makeRequest(ctx context.Context, serverAddr, name string){
	span, ctx := opentracing.IntroduceSpan(ctx, "server-1_to_server-2")
	defer span.Finish()

	req, err := http.NewRequest(http.MethodPost, serverAddr, strings.NewReader(name))
	if err != nil {
		fmt.Printf("http.NewRequest() error : %v", err)
		return
	}

	opentracing.Serialize(ctx, req)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("c.Do() error : %v", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error : %v", err)
		return
	}

	fmt.Println(string(data))
}