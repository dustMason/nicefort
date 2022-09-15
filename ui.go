package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"time"
)

type Action int

const (
	Up Action = iota
	Right
	Down
	Left
	Disconnect
)

type UIModel struct {
	world      *World
	playerID   string
	playerName string
	width      int
	height     int
	quitting   bool
	keys       keyMap
	lastKey    string
	help       help.Model
	chatInput  textinput.Model
}

func NewUIModel(w *World, playerID, playerName string, width, height int) UIModel {
	ti := textinput.New()
	ti.Placeholder = "chat"
	ti.CharLimit = 156
	ti.Width = 27
	return UIModel{
		world:      w,
		playerID:   playerID,
		playerName: playerName,
		width:      width,
		height:     height,
		keys:       keys,
		help:       help.New(),
		chatInput:  ti,
	}
}

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Help      key.Binding
	FocusChat key.Binding
	Enter     key.Binding
	Esc       key.Binding
	Quit      key.Binding
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
	FocusChat: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "focus chat input"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m UIModel) Init() tea.Cmd {
	return doTick()
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		return m, doTick()
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}
	if m.chatInput.Focused() {
		return m.handleChatModeMessage(msg)
	}
	return m.handleNormalModeMessage(msg)
}

func (m UIModel) handleChatModeMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			if m.chatInput.Focused() {
				m.world.EmitEvent(fmt.Sprintf("%s: %s", m.playerName, m.chatInput.Value()))
				m.chatInput.SetValue("")
				m.chatInput.Blur()
			}
		case key.Matches(msg, m.keys.Esc):
			m.chatInput.SetValue("")
			m.chatInput.Blur()
		}
	}
	m.chatInput, cmd = m.chatInput.Update(msg)
	return m, cmd

}

func (m UIModel) handleNormalModeMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case key.Matches(msg, m.keys.FocusChat):
			if !m.chatInput.Focused() {
				m.chatInput.Focus()
			}
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		}
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
			Border(lipgloss.RoundedBorder()) // .Width(60).Height(30)

	feedBoxStyle = lipgloss.NewStyle().
			Inherit(dialogBoxStyle).Width(20) // .Height(27)

	chatInputStyle = lipgloss.NewStyle().
			Inherit(dialogBoxStyle).Height(1).Width(20)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"}).
			Width(84)

	docStyle = lipgloss.NewStyle().Padding(0)
)

func (m UIModel) View() string {
	mapWidth := m.width - 24
	mapHeight := m.height - 4

	// local copy, because Width/Height mutate it. this avoids `concurrent map write` panics
	mbStyle := lipgloss.NewStyle().Inherit(mapBoxStyle).Width(mapWidth).Height(mapHeight)

	doc := strings.Builder{}
	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			mbStyle.Render(m.world.Render(m.playerID, m.playerName, mapWidth, mapHeight)),
			lipgloss.JoinVertical(
				lipgloss.Left,
				feedBoxStyle.Render(strings.Join(m.world.events, "\n\n")),
				chatInputStyle.Render(m.chatInput.View()),
			),
		),
		statusBarStyle.Render(m.lastKey),
	)
	doc.WriteString(ui)
	return docStyle.Render(doc.String())
}
