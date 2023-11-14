package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct {
	ListenAddr string
	Api  *ApiServer
}

func NewRouter(l string, a *ApiServer) *Router {
  return &Router{
    ListenAddr: l,
    Api: a,
  }
}

func (r *Router) Run() {
  m := mux.NewRouter() 

  m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "hello world")
  })
  m.HandleFunc("/api/register", r.Api.HandleRegister)
  m.HandleFunc("/api/login", r.Api.HandleLogin)

  log.Printf("server running in port %s", r.ListenAddr)
  log.Fatal(http.ListenAndServe(r.ListenAddr, m))
}

