package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	host = "localhost"
	port = 23234
)

type Action int

const (
	Up Action = iota
	Right
	Down
	Left
	Disconnect
)

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Help  key.Binding
	Quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right}, // first column
		{k.Help, k.Quit},                // second column
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type model struct {
	world      *World
	quitting   bool
	keys       keyMap
	lastKey    string
	help       help.Model
	term       string
	playerID   string
	playerName string
}

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return doTick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.lastKey = "↑"
			m.world.ApplyCommand(Up, m.playerID, m.playerName)
		case key.Matches(msg, m.keys.Down):
			m.lastKey = "↓"
			m.world.ApplyCommand(Down, m.playerID, m.playerName)
		case key.Matches(msg, m.keys.Left):
			m.lastKey = "←"
			m.world.ApplyCommand(Left, m.playerID, m.playerName)
		case key.Matches(msg, m.keys.Right):
			m.lastKey = "→"
			m.world.ApplyCommand(Right, m.playerID, m.playerName)
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		}
	case TickMsg:
		return m, doTick()
	}
	return m, nil
}

var (
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	mapBoxStyle = lipgloss.NewStyle().
			Inherit(dialogBoxStyle).
			Width(60).Height(30)

	feedBoxStyle = lipgloss.NewStyle().
			Inherit(dialogBoxStyle).
			Width(20).Height(30)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"}).
			Width(84)

	docStyle = lipgloss.NewStyle().Padding(0)
)

func (m model) View() string {
	doc := strings.Builder{}
	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			mapBoxStyle.Render(m.world.Render(m.playerID, m.playerName, 30, 29)),
			feedBoxStyle.Render(strings.Join(m.world.events, "\n\n")),
		),
		statusBarStyle.Render(m.lastKey),
	)
	doc.WriteString(ui)
	return docStyle.Render(doc.String())
}

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

	return &Server{
		world: w,
		ssh:   s,
	}
}

func DisconnectHandlerMiddleware(w *World) wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			pubKey := string(gossh.MarshalAuthorizedKey(s.PublicKey()))
			playerName := s.User()
			sh(s)
			w.ApplyCommand(Disconnect, pubKey, playerName)
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
		m := model{
			term:       pty.Term,
			keys:       keys,
			help:       help.New(),
			world:      w,
			playerID:   pubKey,
			playerName: s.User(),
		}
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
