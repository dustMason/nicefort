package world

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustmason/nicefort/events"
	"github.com/dustmason/nicefort/util"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const NPCActivationRadius = 10 // approx. distance from any player where NPCs take turns

type Coord struct {
	X, Y int
}

func (c Coord) GetXY() (int, int) {
	return c.X, c.Y
}

type World struct {
	sync.RWMutex
	W, H       int
	wMap       []location         // the actual map of tiles
	players    map[string]*entity // map of player id => entity that points to that player
	activeNPCs []*entity
	events     *events.EventList
	onEvent    map[string]func(string) // map of player id => chat callback
}

// SizeX SizeY IsPassable and OOB to satisfy the dmap interface
func (w *World) SizeX() int {
	return w.W
}

func (w *World) SizeY() int {
	return w.H
}

func (w *World) IsPassable(x int, y int) bool {
	return w.walkable(x, y)
}

func (w *World) OOB(x int, y int) bool {
	return !w.InBounds(x, y)
}

type location []*entity

func removeEntity(l location, e *entity) location {
	ret := make(location, 0)
	for _, ent := range l {
		if ent != e {
			ret = append(ret, ent)
		}
	}
	return ret
}

func NewWorld(size int) *World {
	w := &World{
		W:       size,
		H:       size,
		players: make(map[string]*entity),
		wMap:    GenerateOverworld(size),
		events:  events.NewEventList(100),
		onEvent: make(map[string]func(string)),
	}
	go w.runTicker()
	return w
}

func (w *World) tick(t time.Time) {
	for _, e := range w.activeNPCs {
		e.npc.Tick(t, w, e)
	}
	for _, e := range w.players {
		e.player.Tick(t)
	}
}

func (w *World) runTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for t := range ticker.C {
		w.tick(t)
	}
}

func (w *World) MovePlayer(dx, dy int, playerID string) {
	e := w.getPlayer(playerID)
	w.movePlayer(e, dx, dy)
}

func (w *World) MoveNPC(x, y int, e *entity) {
	w.Lock()
	defer w.Unlock()
	if w.InBounds(x, y) && w.walkable(x, y) && !w.occupied(x, y) { // todo some NPCs can move over different types of terrain
		newIndex := w.index(x, y)
		oldIndex := w.index(e.npc.loc.X, e.npc.loc.Y)
		w.wMap[newIndex] = append(w.wMap[newIndex], e)
		w.wMap[oldIndex] = removeEntity(w.wMap[oldIndex], e)
		e.npc.loc = Coord{x, y}
	}
}

func (w *World) getOrCreatePlayer(playerID, playerName string) *entity {
	w.Lock()
	defer w.Unlock()
	e, ok := w.players[playerID]
	if !ok {
		x, y, _ := w.randomAvailableCoord()
		e = NewPlayer(playerID, Coord{x, y})
		w.players[playerID] = e
	}
	e.player.name = playerName
	x, y := e.player.GetLocation()
	if !w.isPlayerAtLocation(e, x, y) {
		// ensure that a `wMap` entry exists. (this might be a reconnecting player)
		i := w.index(x, y)
		w.wMap[i] = append(w.wMap[i], e)
		e.player.See(w)
		w.refreshActiveNPCs()
	}
	return e
}

func (w *World) refreshActiveNPCs() {
	// from each player, grab all NPCs within a boundary and make sure they are all in activeNPCs
	// set any remaining to inactive
	found := make(map[*entity]struct{})
	for _, e := range w.players {
		px, py := e.player.GetLocation()
		x1 := util.ClampedInt(px-NPCActivationRadius, 0, w.W-1)
		y1 := util.ClampedInt(py-NPCActivationRadius, 0, w.H-1)
		x2 := util.ClampedInt(px+NPCActivationRadius, 0, w.W-1)
		y2 := util.ClampedInt(py+NPCActivationRadius, 0, w.H-1)
		ix := x1
		iy := y1
		for iy < y2 {
			for ix < x2 {
				for _, ent := range w.location(ix, iy) {
					if ent.npc != nil {
						found[ent] = struct{}{}
					}
				}
				ix++
			}
			ix = x1
			iy++
		}
	}
	newActiveNPCs := make([]*entity, 0)
	for e := range found {
		newActiveNPCs = append(newActiveNPCs, e)
	}
	w.activeNPCs = newActiveNPCs
}

func (w *World) getPlayer(playerID string) *entity {
	w.Lock()
	defer w.Unlock()
	e, _ := w.players[playerID]
	return e
}

func (w *World) isPlayerAtLocation(e *entity, x, y int) bool {
	for _, ent := range w.location(x, y) {
		if ent == e {
			return true
		}
	}
	return false
}

func (w *World) movePlayer(e *entity, dx, dy int) {
	nx, ny := e.player.GetLocation()
	nx += dx
	ny += dy
	now := time.Now()
	if !e.player.CanMove(now) {
		return
	}
	w.Lock()
	defer w.Unlock()
	newIndex := w.index(nx, ny)
	if w.InBounds(nx, ny) {
		if ent, ok := w.pickupable(nx, ny); ok {
			took := e.player.PickUp(ent.item, ent.quantity)
			ent.quantity -= took
			if ent.quantity == 0 {
				w.wMap[newIndex] = removeEntity(w.wMap[newIndex], ent)
			}
			return
		}
		if ent, ok := w.attackable(nx, ny); ok {
			damage, dead, _ := ent.npc.Attacked(e, 10)
			if dead {
				e.player.Event(events.Success, fmt.Sprintf("You killed the %s", ent.npc.Name))
				i := w.index(ent.npc.loc.X, ent.npc.loc.Y)
				w.wMap[i] = removeEntity(w.wMap[i], ent)
				// todo handle drops
			} else {
				e.player.Event(events.Success, fmt.Sprintf("You hit the %s for %d", ent.npc.Name, damage))
			}
			return
		}
		if w.walkable(nx, ny) && !w.occupied(nx, ny) {
			oldX, oldY := e.player.GetLocation()
			oldIndex := w.index(oldX, oldY)
			w.wMap[newIndex] = append(w.wMap[newIndex], e)
			w.wMap[oldIndex] = removeEntity(w.wMap[oldIndex], e)
			e.player.SetLocation(nx, ny, now)
			e.player.See(w)
			w.refreshActiveNPCs()
			return
		}
	}
}

var blackSpace = tileMap[Environment][Space][0]
var memColor = "#444444"
var memStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(memColor))

func (w *World) RenderMap(playerID, playerName string, vw, vh int) string {
	vw = vw / 2 // each tile is 2 chars wide
	ply := w.getOrCreatePlayer(playerID, playerName).player
	x, y := ply.GetLocation()
	seen := ply.AllVisited()

	w.RLock()
	defer w.RUnlock()
	var b strings.Builder
	b.Grow(vw*vh*2 + vh) // *2 because double-width chars and +vh because line-breaks
	left := x - vw/2
	right := x + vw/2
	top := y - vh/2
	bottom := y + vh/2
	ix := left
	iy := top
	for iy < bottom {
		for ix < right {
			var ent *entity
			if !w.InBounds(ix, iy) {
				b.WriteString(blackSpace)
			} else {
				ic := Coord{ix, iy}
				inView, dist := ply.CanSee(ix, iy)
				memString, inPastView := seen[ic]
				if inPastView && !inView {
					b.WriteString(memString)
				} else if inView {
					loc := w.location(ix, iy)
					ent = loc[len(loc)-1]
					bg := loc[0]
					b.WriteString(
						lipgloss.NewStyle().
							Foreground(lipgloss.Color(ent.ForegroundColor(dist))).
							Background(lipgloss.Color(bg.BackgroundColor(dist))).
							Render(ent.String()),
					)
				} else { // not in past or current view
					b.WriteString(blackSpace)
				}
			}
			ix++
		}
		ix = left
		b.WriteString("\n")
		iy++
	}
	return b.String()
}

func (w *World) RenderForMemory(x, y int) string {
	loc := w.location(x, y)
	// iterate backwards to get topmost memorable entity first
	for i := len(loc) - 1; i >= 0; i-- {
		if loc[i].Memorable() {
			return memStyle.Render(loc[i].String())
		}
	}
	return blackSpace
}

func (w *World) randomAvailableCoord() (int, int, error) {
	tries := 1000
	for tries > 0 {
		x := rand.Intn(w.W)
		y := rand.Intn(w.H)
		if w.walkable(x, y) && !w.occupied(x, y) {
			return x, y, nil
		}
		tries--
	}
	return 0, 0, errors.New("couldn't place random coord")
}

func (w *World) InBounds(x, y int) bool {
	return x >= 0 && x < w.W && y >= 0 && y < w.H
}

func (w *World) IsOpaque(x, y int) bool {
	for _, e := range w.location(x, y) {
		if !e.SeeThrough() {
			return true
		}
	}
	return false
}

func (w *World) walkable(x, y int) bool {
	for _, e := range w.location(x, y) {
		if !e.Walkable() {
			return false
		}
	}
	return true
}

func (w *World) occupied(x, y int) bool {
	for _, e := range w.location(x, y) {
		if e.class != Environment {
			return true
		}
	}
	return false
}

func (w *World) pickupable(x, y int) (*entity, bool) {
	for _, e := range w.location(x, y) {
		if e.Pickupable() {
			return e, true
		}
	}
	return nil, false
}

func (w *World) attackable(x int, y int) (*entity, bool) {
	for _, e := range w.location(x, y) {
		if e.Attackable() {
			return e, true
		}
	}
	return nil, false
}

func (w *World) index(x, y int) int {
	return y*w.W + x
}

func (w *World) location(x, y int) location {
	return w.wMap[w.index(x, y)]
}

func (w *World) coordinates(i int) (int, int) {
	x := i % w.W
	y := i / w.H
	return x, y
}

func (w *World) OnEvent(playerID string, f func(string)) {
	w.onEvent[playerID] = f
}

func (w *World) Event(kind events.Class, message string) {
	w.events.Add(kind, message)
	s := w.events.Render()
	for _, f := range w.onEvent {
		f(s)
	}
}

func (w *World) Chat(kind events.Class, subject, message string) {
	w.events.AddWithSubject(kind, message, subject)
	s := w.events.Render()
	for _, f := range w.onEvent {
		f(s)
	}
}

func (w *World) disconnectPlayer(e *entity) {
	w.Lock()
	defer w.Unlock()
	x, y := e.player.GetLocation()
	i := w.index(x, y)
	w.wMap[i] = removeEntity(w.wMap[i], e)
}

// todo refactor this to return some value type (map of string[string]?) or a struct
func (w *World) RenderPlayerSidebar(id string, name string) string {
	var b strings.Builder
	e := w.getOrCreatePlayer(id, name)
	b.WriteString(name + "\n")
	b.WriteString(fmt.Sprintf("Pack: %.1f / %d\n", e.player.carrying, int(e.player.maxCarry)))
	b.WriteString(fmt.Sprintf("Health: %d / %d\n", e.player.health, e.player.maxHealth))
	b.WriteString(fmt.Sprintf("Cash: $%d\n", e.player.money))
	b.WriteString(fmt.Sprintf("Hunger: %.3f\n", e.player.hunger))
	b.WriteString("\n")

	for _, ee := range w.players {
		if ee == e {
			continue
		}
		b.WriteString(
			compassIndicator(e.player.loc.X, e.player.loc.Y, ee.player.loc.X, ee.player.loc.Y) + " " + ee.player.name + "\n",
		)
	}

	b.WriteString("\n")
	for _, ee := range w.activeNPCs {
		b.WriteString(
			compassIndicator(e.player.loc.X, e.player.loc.Y, ee.npc.loc.X, ee.npc.loc.Y) + " " + ee.npc.Name + "\n",
		)
	}

	// todo also get a list of the items the player can see and render compass items for those

	return b.String()
}

var arrows = map[[2]int]string{
	[2]int{0, 0}:   "•",
	[2]int{1, 1}:   "↘",
	[2]int{1, -1}:  "↗",
	[2]int{-1, 1}:  "↙",
	[2]int{-1, -1}: "↖",
	[2]int{0, -1}:  "↑",
	[2]int{0, 1}:   "↓",
	[2]int{-1, 0}:  "←",
	[2]int{1, 0}:   "→",
}

// compassIndicator returns a string like "↖ 30"
func compassIndicator(fromX, fromY, toX, toY int) string {
	dx := toX - fromX
	dy := toY - fromY
	deltaX := util.ClampedInt(dx, -1, 1)
	deltaY := util.ClampedInt(dy, -1, 1)
	arrow, ok := arrows[[2]int{deltaX, deltaY}]
	out := ""
	if ok {
		out = arrow + " "
	}
	return out + strconv.Itoa(util.AbsInt(dx)+util.AbsInt(dy))
}

func (w *World) ActivateItem(playerID string, inventoryIndex int) {
	// todo get the item from the player's inventory and activate it!
	// call a method on the player to consumer the item if the activate function returns true
}

func (w *World) PlayerInventory(playerID string) []*InventoryItem {
	e := w.getPlayer(playerID)
	return e.player.Inventory()
}

func (w *World) AvailableRecipes(playerID string) []Recipe {
	e := w.getPlayer(playerID)
	return AvailableRecipes(e.player.inventoryMap, e, w)
}

func (w *World) DoRecipe(playerID string, r Recipe) bool {
	e := w.getPlayer(playerID)
	ok, newInv := r.Do(e.player.inventoryMap, e, w)
	if ok {
		e.player.ReplaceInventory(newInv)
		return true
	}
	return false
}

func (w *World) DisconnectPlayer(playerID string) {
	e := w.getPlayer(playerID)
	w.disconnectPlayer(e)
	w.Event(events.Warning, fmt.Sprintf("%s left.", e.player.name))
	w.Lock()
	defer w.Unlock()
	delete(w.onEvent, playerID)
	w.refreshActiveNPCs()
}

func (w *World) RenderPlayerEvents(playerID string) string {
	e := w.getPlayer(playerID)
	return e.player.Events()
}
