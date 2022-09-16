package main

func main() {
	w := NewWorld(500)
	s := NewServer(w)
	s.Listen()
}
