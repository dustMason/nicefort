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
	W, H     int
	entities map[Coord]*entity  // the actual map of tiles
	players  map[string]*entity // map of player id => entity that points to that player
	events   []string
}

func NewWorld() *World {
	w := &World{
		W:        50,
		H:        50,
		players:  make(map[string]*entity),
		entities: GenerateMap(50, 50),
	}
	// TODO create a bunch of entities and populate the map
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
	if _, ok := w.entities[e.player.loc]; !ok {
		// ensure that an `entities` entry exists. (this might be a reconnecting player)
		w.entities[e.player.loc] = e
		e.player.See(w)
	}
	return e
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
	if w.InBounds(newLoc.X, newLoc.Y) && w.available(newLoc) {
		w.Lock()
		defer w.Unlock()
		w.entities[newLoc] = e
		delete(w.entities, e.player.loc)
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
					if e, ok := w.entities[ic]; ok {
						ent = e
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
	if e, ok := w.entities[ic]; ok {
		return memStyle.Render(e.String())
	}
	return tileMap[Environment][Space]
}

func (w *World) randomAvailableCoord() (Coord, error) {
	tries := 1000
	for tries > 0 {
		c := Coord{rand.Intn(w.W), rand.Intn(w.H)}
		if _, ok := w.entities[c]; !ok {
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
	loc := Coord{x, y}
	if e, ok := w.entities[loc]; ok {
		return !e.SeeThrough()
	}
	return false
}

func (w *World) available(loc Coord) bool {
	w.RLock()
	defer w.RUnlock()
	_, ok := w.entities[loc]
	return !ok
}

func (w *World) EmitEvent(message string) {
	w.events = append(w.events, message)
}

func (w *World) disconnectPlayer(playerID string, loc Coord) {
	w.Lock()
	defer w.Unlock()
	delete(w.entities, loc)
}
