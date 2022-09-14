package main

func main() {
	w := NewWorld()
	s := NewServer(w)
	s.Listen()
}
