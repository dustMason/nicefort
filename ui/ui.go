package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustmason/nicefort/events"
	"github.com/dustmason/nicefort/world"
	"github.com/muesli/reflow/wordwrap"
	"strconv"
	"strings"
	"time"
)

type Mode int
type InventoryMode int

const (
	Map Mode = iota
	Inventory
)

const (
	InventoryList InventoryMode = iota
	RecipeList
)

type UIModel struct {
	mode          Mode
	world         *world.World
	playerID      string
	playerName    string
	width         int
	height        int
	quitting      bool
	keys          keyMap
	lastKey       string
	help          help.Model
	chat          *viewport.Model
	chatInput     textinput.Model
	inventory     table.Model
	recipes       list.Model
	inventoryMode InventoryMode
}

func NewUIModel(w *world.World, playerID, playerName string, width, height int) UIModel {
	ti := textinput.New()
	ti.Placeholder = "chat"
	ti.CharLimit = 240
	ti.Width = 17

	chat := viewport.New(20, height-5) // -5 for chatInput
	w.OnEvent(playerID, func(events string) {
		chat.SetContent(wordwrap.String(events, 20))
		chat.GotoBottom()
	})

	return UIModel{
		mode:          Map,
		world:         w,
		playerID:      playerID,
		playerName:    playerName,
		width:         width,
		height:        height,
		keys:          keys,
		help:          help.New(),
		chat:          &chat,
		chatInput:     ti,
		inventory:     table.New(),
		inventoryMode: InventoryList,
	}
}

type keyMap struct {
	Up             key.Binding
	Down           key.Binding
	Left           key.Binding
	Right          key.Binding
	Help           key.Binding
	FocusChat      key.Binding
	FocusInventory key.Binding
	Enter          key.Binding
	Tab            key.Binding
	Esc            key.Binding
	Quit           key.Binding
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
	FocusInventory: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "show inventory browser"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
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
		m.chat.Height = msg.Height - 5
	}
	if m.chatInput.Focused() {
		return m.handleChatModeMessage(msg)
	}
	if m.mode == Inventory {
		return m.handleInventoryModeMessage(msg)
	}
	return m.handleMapModeMessage(msg)
}

func (m UIModel) handleChatModeMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Enter):
			if m.chatInput.Focused() {
				m.world.Chat(events.Info, m.playerName, m.chatInput.Value())
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

func (m UIModel) handleMapModeMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.lastKey = "↑"
			m.world.MovePlayer(0, -1, m.playerID)
		case key.Matches(msg, m.keys.Down):
			m.lastKey = "↓"
			m.world.MovePlayer(0, 1, m.playerID)
		case key.Matches(msg, m.keys.Left):
			m.lastKey = "←"
			m.world.MovePlayer(-1, 0, m.playerID)
		case key.Matches(msg, m.keys.Right):
			m.lastKey = "→"
			m.world.MovePlayer(1, 0, m.playerID)
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.FocusChat):
			if !m.chatInput.Focused() {
				m.chatInput.Focus()
			}
		case key.Matches(msg, m.keys.FocusInventory):
			m.inventory = m.createInventoryTable()
			m.recipes = m.createRecipeList()
			m.mode = Inventory
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m UIModel) handleInventoryModeMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Esc):
			if m.recipes.FilterState() == list.Filtering {
				break
			}
			m.mode = Map
			return m, nil
		case key.Matches(msg, m.keys.Tab):
			if m.inventoryMode == InventoryList {
				m.inventoryMode = RecipeList
				m.inventory.Blur()
			} else {
				m.inventoryMode = InventoryList
				m.inventory.Focus()
			}
		case key.Matches(msg, m.keys.Enter):
			if m.inventoryMode == InventoryList {
				i, _ := strconv.Atoi(m.inventory.SelectedRow()[0])
				m.world.ActivateItem(m.playerID, i)
				m.mode = Map
			} else {
				if m.recipes.FilterState() == list.Filtering {
					break
				}
				if i, ok := m.recipes.SelectedItem().(recipeListItem); ok {
					if fok, selectedRecipe := world.FindRecipe(i.id); fok {
						if m.world.DoRecipe(m.playerID, selectedRecipe) {
							m.inventory.SetRows(m.createInventoryTableRows())
						}
						cmd = m.recipes.SetItems(m.createRecipeListItems())
						var cmd2 tea.Cmd
						m.recipes, cmd2 = m.recipes.Update(TickMsg{})
						return m, tea.Batch(cmd, cmd2)
					}
				}
			}
		}
	}
	if m.inventory.Focused() {
		m.inventory, cmd = m.inventory.Update(msg)
	} else {
		m.recipes, cmd = m.recipes.Update(msg)
	}
	return m, cmd
}

func (m UIModel) createInventoryTable() table.Model {
	// todo fix these widths to handle flexible viewport width
	columns := []table.Column{
		{Title: "Item", Width: 45},
		{Title: "Qty", Width: 5},
		{Title: "Weight", Width: 8},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(m.createInventoryTableRows()),
		table.WithFocused(true),
		// table.WithHeight(7),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
}

func (m UIModel) createInventoryTableRows() []table.Row {
	ii := m.world.PlayerInventory(m.playerID)
	rows := make([]table.Row, len(ii))
	for ind, i := range ii {
		rows[ind] = table.Row{i.Item.Name, strconv.Itoa(i.Quantity), fmt.Sprintf("%.1f", i.Weight())}
	}
	return rows
}

type recipeListItem struct {
	title       string
	description string
	id          int
}

func (i recipeListItem) Title() string       { return i.title }
func (i recipeListItem) Description() string { return i.description }
func (i recipeListItem) FilterValue() string { return i.title }

func (m UIModel) createRecipeList() list.Model {
	d := list.NewDefaultDelegate()
	lm := list.New(m.createRecipeListItems(), d, 50, 20)
	lm.Title = "Crafting Recipes"
	lm.SetShowHelp(false)
	lm.SetStatusBarItemName("recipe", "recipes")
	lm.DisableQuitKeybindings()
	return lm
}

func (m UIModel) createRecipeListItems() []list.Item {
	items := make([]list.Item, 0)
	for _, r := range m.world.AvailableRecipes(m.playerID) {
		items = append(items, recipeListItem{title: r.Result.Name, description: r.Description, id: r.ID})
	}
	return items
}

var (
	borderedBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#874BFD")).
				Padding(0).
				BorderTop(true).
				BorderLeft(true).
				BorderRight(true).
				BorderBottom(true)

	playerInfoStyle   = lipgloss.NewStyle().Width(20)
	playerEventsStyle = lipgloss.NewStyle().Height(4)
	chatPaneStyle     = lipgloss.NewStyle().Width(20)
	chatInputStyle    = lipgloss.NewStyle().Inherit(borderedBoxStyle).Height(1).Width(20)
	statusBarStyle    = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
				Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})
	docStyle = lipgloss.NewStyle().Padding(0)
)

func (m UIModel) View() string {
	mainWidth := m.width - 20 - 2 - 20 - 2 // minus both sidebars and borders
	mainHeight := m.height - 4 - 4

	// local copy, because Width/Height mutate it. this avoids `concurrent map write` panics
	piStyle := lipgloss.NewStyle().Inherit(playerInfoStyle).Height(m.height - 1) // -1 for statusbar
	peStyle := lipgloss.NewStyle().Inherit(playerEventsStyle).Width(mainWidth).MaxHeight(4)
	mainStyle := lipgloss.NewStyle().Inherit(borderedBoxStyle).Width(mainWidth).Height(mainHeight)
	sbStyle := lipgloss.NewStyle().Inherit(statusBarStyle).Width(m.width)
	inventoryPaneStyle := lipgloss.NewStyle().Width(mainWidth / 2).Height(mainHeight)
	recipePaneStyle := lipgloss.NewStyle().Width(mainWidth / 2).Height(mainHeight)
	if m.inventoryMode == InventoryList {
		recipePaneStyle.Faint(true)
	} else {
		inventoryPaneStyle.Faint(true)
	}

	var mainContents string
	if m.mode == Map {
		mainContents = mainStyle.Render(m.world.RenderMap(m.playerID, m.playerName, mainWidth, mainHeight))
	} else if m.mode == Inventory {
		mainContents = lipgloss.JoinHorizontal(
			lipgloss.Top,
			inventoryPaneStyle.Render(
				m.recipes.Styles.Title.Render("Inventory")+"\n\n"+m.inventory.View(),
			),
			recipePaneStyle.Render(m.recipes.View()),
		)
	}

	doc := strings.Builder{}
	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			piStyle.Render(m.world.RenderPlayerSidebar(m.playerID, m.playerName)),
			lipgloss.JoinVertical(
				lipgloss.Left,
				peStyle.Render(m.world.RenderPlayerEvents(m.playerID)),
				mainContents,
			),
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.chat.View(),
				chatInputStyle.Render(m.chatInput.View()),
			),
		),
		sbStyle.Render(m.lastKey), // todo use status bar for short help text
	)
	doc.WriteString(ui)
	return docStyle.Render(doc.String())
}
