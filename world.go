package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

const Space = "."

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
	entities map[Coord]*entity
	players  map[string]*entity // map of player id => entity that points to that player
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

func (w *World) ApplyCommand(a Action, playerID string) {
	e := w.getOrCreatePlayer(playerID)
	switch a {
	case Up, Right, Down, Left:
		w.movePlayer(e, a)
	default:
		fmt.Println("unknown action:", a)
	}
}

func (w *World) getOrCreatePlayer(playerID string) *entity {
	w.Lock()
	defer w.Unlock()
	e, ok := w.players[playerID]
	if !ok {
		c, _ := w.randomAvailableCoord()
		e = NewPlayer(playerID, c)
		w.players[playerID] = e
		w.entities[c] = e
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
	if w.inBounds(newLoc) && w.available(newLoc) {
		w.Lock()
		defer w.Unlock()
		w.entities[newLoc] = e
		delete(w.entities, e.player.loc)
		e.player.loc = newLoc
	}
}

// todo World needs it's own ticker, running in a loop, that moves NPEs

// Render accepts a playerID and vw,vh viewport size
// it renders a section of the map that player can see
func (w *World) Render(playerID string, vw, vh int) string {
	c := w.getOrCreatePlayer(playerID).player.loc
	x := c.X
	y := c.Y

	var b strings.Builder
	b.Grow(vw*vh + vh)
	left := x - vw/2
	right := x + vw/2
	top := y - vh/2
	bottom := y + vh/2

	w.RLock()
	defer w.RUnlock()
	ix := left
	iy := top
	for iy < bottom {
		for ix < right {
			if ix < 0 || iy < 0 {
				b.WriteString(" ")
			} else {
				if e, ok := w.entities[Coord{ix, iy}]; ok {
					b.WriteString(e.String())
				} else {
					b.WriteString(Space)
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

func (w *World) randomAvailableCoord() (Coord, error) {
	fmt.Println("generating random coord")
	tries := 1000
	for tries > 0 {
		c := Coord{rand.Intn(w.W), rand.Intn(w.H)}
		if _, ok := w.entities[c]; !ok {
			fmt.Println("found", c)
			return c, nil
		}
		tries--
	}
	return Coord{}, errors.New("couldn't place random coord")
}

func (w *World) inBounds(loc Coord) bool {
	return loc.X >= 0 && loc.X < w.W && loc.Y >= 0 && loc.Y < w.H
}

func (w *World) available(loc Coord) bool {
	w.RLock()
	defer w.RUnlock()
	_, ok := w.entities[loc]
	return !ok
}
