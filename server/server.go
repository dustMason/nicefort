package server

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/dustmason/nicefort/ui"
	"github.com/dustmason/nicefort/world"
	"github.com/gliderlabs/ssh"
	"github.com/muesli/termenv"
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
	world *world.World
}

func NewServer(w *world.World) *Server {
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			bm.MiddlewareWithColorProfile(teaHandler(w), termenv.TrueColor),
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

// DisconnectHandlerMiddleware makes sure the world gets cleaned up when a player disconnects.
// their player instance persists so when they reconnect they can resume.
func DisconnectHandlerMiddleware(w *world.World) wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			pubKey := string(gossh.MarshalAuthorizedKey(s.PublicKey()))
			sh(s)
			w.DisconnectPlayer(pubKey)
		}
	}
}

func teaHandler(w *world.World) bm.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil, nil
		}
		pubKey := string(gossh.MarshalAuthorizedKey(s.PublicKey()))
		m := ui.NewUIModel(w, pubKey, s.User(), pty.Window.Width, pty.Window.Height)
		w.PlayerJoin(pubKey, s.User())
		w.OnPlayerDeath(pubKey, func() {
			_ = s.Exit(0)
		})
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
