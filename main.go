package main

import (
	"log"
	"time"
)

const (
  Domain = "localhost"
  Path = "/"
  CookieLifeTime = time.Hour * 24 * 30
  serverPort = ":3000"
  JwtSecret = "abc123"
)

func main() {
  store, err := NewPostgresStore()
  if err != nil {
    log.Fatal(err)
    return
  }
  if err = store.Init(); err != nil {
    log.Fatal(err)
    return
  }

  api := NewApiServer(store)
  router := NewRouter(serverPort, api)

  router.Run()
}
