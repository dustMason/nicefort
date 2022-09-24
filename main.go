package main

import (
	"fmt"
	"github.com/dustmason/nicefort/server"
	"github.com/dustmason/nicefort/world"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		fmt.Println(http.ListenAndServe(":6060", nil))
	}()
	w := world.NewWorld(600)
	s := server.NewServer(w)
	s.Listen()
}
