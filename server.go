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
	world    *World
	quitting bool
	keys     keyMap
	lastKey  string
	help     help.Model
	term     string
	playerID string
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
			m.world.ApplyCommand(Up, m.playerID)
		case key.Matches(msg, m.keys.Down):
			m.lastKey = "↓"
			m.world.ApplyCommand(Down, m.playerID)
		case key.Matches(msg, m.keys.Left):
			m.lastKey = "←"
			m.world.ApplyCommand(Left, m.playerID)
		case key.Matches(msg, m.keys.Right):
			m.lastKey = "→"
			m.world.ApplyCommand(Right, m.playerID)
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

func (m model) View() string {
	doc := strings.Builder{}
	world := dialogBoxStyle.Render(m.world.Render(m.playerID, 60, 30))
	doc.WriteString(world)
	return docStyle.Render(doc.String())
}

var (
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	// statusNugget = lipgloss.NewStyle().
	// 		Foreground(lipgloss.Color("#FFFDF5")).
	// 		Padding(0, 1)
	//
	// statusBarStyle = lipgloss.NewStyle().
	// 		Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
	// 		Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})
	//
	// statusStyle = lipgloss.NewStyle().
	// 		Inherit(statusBarStyle).
	// 		Foreground(lipgloss.Color("#FFFDF5")).
	// 		Background(lipgloss.Color("#FF5F87")).
	// 		Padding(0, 1).
	// 		MarginRight(1)
	//
	// encodingStyle = statusNugget.Copy().
	// 		Background(lipgloss.Color("#A550DF")).
	// 		Align(lipgloss.Right)
	//
	// statusText = lipgloss.NewStyle().Inherit(statusBarStyle)
	//
	// fishCakeStyle = statusNugget.Copy().Background(lipgloss.Color("#6124DF"))

	docStyle = lipgloss.NewStyle().Padding(0)
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
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	return &Server{
		world: w,
		ssh:   s,
	}
}

func teaHandler(w *World) bm.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil, nil
		}
		m := model{
			term:     pty.Term,
			keys:     keys,
			help:     help.New(),
			world:    w,
			playerID: s.User(),
		}
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
