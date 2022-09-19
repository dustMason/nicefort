package main

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	host = ""
	port = 23234
)

type Server struct {
	ssh   *ssh.Server
	world *World
}

func NewServer(w *World) *Server {
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			bm.Middleware(teaHandler(w)),
			DisconnectHandlerMiddleware(w),
			lm.Middleware(),
		),
		ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true // allow all keys
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return &Server{world: w, ssh: s}
}

func DisconnectHandlerMiddleware(w *World) wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			pubKey := string(gossh.MarshalAuthorizedKey(s.PublicKey()))
			playerName := s.User()
			sh(s)
			w.ApplyCommand(Disconnect, pubKey)
			w.EmitEvent(fmt.Sprintf("%s left.", playerName))
		}
	}
}

func teaHandler(w *World) bm.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil, nil
		}
		pubKey := string(gossh.MarshalAuthorizedKey(s.PublicKey()))
		m := NewUIModel(w, pubKey, s.User(), pty.Window.Width, pty.Window.Height)
		w.EmitEvent(fmt.Sprintf("%s joined.", s.User()))
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}
}

func (s *Server) Listen() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)
	go func() {
		if err := s.ssh.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.ssh.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
