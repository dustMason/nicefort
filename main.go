package main

import (
	"github.com/dustmason/nicefort/server"
	"github.com/dustmason/nicefort/world"
)

func main() {
	w := world.NewWorld(500)
	s := server.NewServer(w)
	s.Listen()
}
