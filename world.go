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

func (c Coord) Add(delta Coord) Coord {
	return Coord{
		X: c.X + delta.X,
		Y: c.Y + delta.Y,
	}
}

type World struct {
	sync.RWMutex
	W, H    int
	wMap    map[Coord]location // the actual map of tiles
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

func NewWorld() *World {
	w := &World{
		W:       500,
		H:       500,
		players: make(map[string]*entity),
		wMap:    GenerateOverworld(500),
	}
	return w
}

func (w *World) ApplyCommand(a Action, playerID, playerName string) {
	e := w.getOrCreatePlayer(playerID, playerName)
	switch a {
	case Up, Right, Down, Left:
		w.movePlayer(e, a)
	case Disconnect:
		w.disconnectPlayer(playerID, e.player.loc)
	default:
		fmt.Println("unknown action:", a)
	}
}

func (w *World) getOrCreatePlayer(playerID, playerName string) *entity {
	w.Lock()
	defer w.Unlock()
	e, ok := w.players[playerID]
	if !ok {
		c, _ := w.randomAvailableCoord()
		e = NewPlayer(playerID, c)
		w.players[playerID] = e
	}
	e.player.name = playerName
	if !w.isPlayerAtLocation(e, e.player.loc) {
		// ensure that a `wMap` entry exists. (this might be a reconnecting player)
		w.wMap[e.player.loc] = append(w.wMap[e.player.loc], e)
		e.player.See(w)
	}
	return e
}

func (w *World) isPlayerAtLocation(e *entity, c Coord) bool {
	loc, ok := w.wMap[c]
	if !ok {
		return false
	}
	for _, ent := range loc {
		if ent == e {
			return true
		}
	}
	return false
}

func (w *World) movePlayer(e *entity, a Action) {
	delta := Coord{}
	switch a {
	case Up:
		delta.Y = -1
	case Right:
		delta.X = 1
	case Down:
		delta.Y = 1
	case Left:
		delta.X = -1
	}
	newLoc := e.player.loc.Add(delta)
	w.Lock()
	defer w.Unlock()
	if w.InBounds(newLoc.X, newLoc.Y) && w.walkable(newLoc) {
		w.wMap[newLoc] = append(w.wMap[newLoc], e)
		w.wMap[e.player.loc] = removeEntity(w.wMap[e.player.loc], e)
		e.player.loc = newLoc
		e.player.See(w)
	}
}

// todo World needs it's own ticker, running in a loop, that moves NPEs

func (w *World) Render(playerID, playerName string, vw, vh int) string {
	vw = vw / 2 // each tile is 2 chars wide
	ply := w.getOrCreatePlayer(playerID, playerName).player
	c := ply.loc
	x := c.X
	y := c.Y
	seen := ply.AllVisited()

	w.RLock()
	defer w.RUnlock()
	var b strings.Builder
	b.Grow(vw*vh + vh)
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
					if loc, ok := w.wMap[ic]; ok {
						ent = loc[len(loc)-1]
					} else {
						ent = &entity{class: Environment, subclass: Floor}
					}
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
	ic := Coord{x, y}
	if loc, ok := w.wMap[ic]; ok {
		return memStyle.Render(loc[len(loc)-1].String())
	}
	return tileMap[Environment][Space]
}

func (w *World) randomAvailableCoord() (Coord, error) {
	tries := 1000
	for tries > 0 {
		c := Coord{rand.Intn(w.W), rand.Intn(w.H)}
		if w.walkable(c) {
			return c, nil
		}
		tries--
	}
	return Coord{}, errors.New("couldn't place random coord")
}

func (w *World) InBounds(x, y int) bool {
	return x >= 0 && x < w.W && y >= 0 && y < w.H
}

func (w *World) IsOpaque(x, y int) bool {
	c := Coord{x, y}
	if loc, ok := w.wMap[c]; ok {
		for _, e := range loc {
			if !e.SeeThrough() {
				return true
			}
		}
	}
	return false
}

func (w *World) walkable(c Coord) bool {
	loc, ok := w.wMap[c]
	if !ok {
		return true
	}
	for _, e := range loc {
		if !e.Walkable() {
			return false
		}
	}
	return true
}

func (w *World) EmitEvent(message string) {
	w.events = append(w.events, message)
}

func (w *World) disconnectPlayer(playerID string, c Coord) {
	ent := w.getOrCreatePlayer(playerID, "")
	w.Lock()
	defer w.Unlock()
	if loc, ok := w.wMap[c]; ok {
		w.wMap[c] = removeEntity(loc, ent)
	}
}
