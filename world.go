package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"math/rand"
	"strings"
	"sync"
)

type Coord struct {
	X, Y int
}

type World struct {
	sync.RWMutex
	W, H int
	// wMap    map[Coord]location // the actual map of tiles
	wMap    []location         // the actual map of tiles
	players map[string]*entity // map of player id => entity that points to that player
	events  []string
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
	}
	return w
}

func (w *World) ApplyCommand(a Action, playerID string) {
	e := w.getPlayer(playerID)
	switch a {
	case Up, Right, Down, Left:
		w.movePlayer(e, a)
	case Disconnect:
		w.disconnectPlayer(e)
	default:
		fmt.Println("unknown action:", a)
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
	if !w.isPlayerAtLocation(e, e.player.loc.X, e.player.loc.Y) {
		// ensure that a `wMap` entry exists. (this might be a reconnecting player)
		i := w.index(e.player.loc.X, e.player.loc.Y)
		w.wMap[i] = append(w.wMap[i], e)
		e.player.See(w)
	}
	return e
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

func (w *World) movePlayer(e *entity, a Action) {
	nx := e.player.loc.X
	ny := e.player.loc.Y
	switch a {
	case Up:
		ny += -1
	case Right:
		nx += 1
	case Down:
		ny += 1
	case Left:
		nx += -1
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
		if w.walkable(nx, ny) {
			oldIndex := w.index(e.player.loc.X, e.player.loc.Y)
			w.wMap[newIndex] = append(w.wMap[newIndex], e)
			w.wMap[oldIndex] = removeEntity(w.wMap[oldIndex], e)
			e.player.loc = Coord{nx, ny}
			e.player.See(w)
			return
		}
	}
}

// todo World needs it's own ticker, running in a loop, that moves NPEs

func (w *World) RenderMap(playerID, playerName string, vw, vh int) string {
	vw = vw / 2 // each tile is 2 chars wide
	ply := w.getOrCreatePlayer(playerID, playerName).player
	c := ply.loc
	x := c.X
	y := c.Y
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
				b.WriteString(tileMap[Environment][Space])
			} else {
				ic := Coord{ix, iy}
				inView := ply.CanSee(ix, iy)
				memString, inPastView := seen[ic]
				if inPastView && !inView {
					b.WriteString(memString)
				} else if inView {
					loc := w.location(ix, iy)
					ent = loc[len(loc)-1]
					b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ent.Color())).Render(ent.String()))
				} else { // not in past or current view
					b.WriteString(tileMap[Environment][Space])
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

var memStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

func (w *World) RenderForMemory(x, y int) string {
	loc := w.location(x, y)
	// iterate backwards to get topmost memorable entity first
	for i := len(loc) - 1; i >= 0; i-- {
		if loc[i].Memorable() {
			return memStyle.Render(loc[i].String())
		}
	}
	return tileMap[Environment][Space]
}

func (w *World) randomAvailableCoord() (int, int, error) {
	tries := 1000
	for tries > 0 {
		x := rand.Intn(w.W)
		y := rand.Intn(w.H)
		if w.walkable(x, y) {
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

func (w *World) pickupable(x, y int) (*entity, bool) {
	for _, e := range w.location(x, y) {
		if e.Pickupable() {
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

func (w *World) EmitEvent(message string) {
	w.events = append(w.events, message)
}

func (w *World) disconnectPlayer(e *entity) {
	w.Lock()
	defer w.Unlock()
	i := w.index(e.player.loc.X, e.player.loc.Y)
	w.wMap[i] = removeEntity(w.wMap[i], e)
}

func (w *World) RenderPlayerSidebar(id string, name string) string {
	var b strings.Builder
	e := w.getOrCreatePlayer(id, name)
	b.WriteString(name + "\n")
	b.WriteString(fmt.Sprintf("Pack: %.1f / %d\n", e.player.carrying, int(e.player.maxCarry)))
	b.WriteString(fmt.Sprintf("Health: %d / %d\n", e.player.health, e.player.maxHealth))
	b.WriteString(fmt.Sprintf("Cash: $%d\n", e.player.money))
	return b.String()
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
