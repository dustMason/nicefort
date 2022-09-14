package main

import (
	"fmt"
	"strings"
	"sync"
)

const Space = "."

type coord struct {
	X, Y int
}

func (c coord) Add(delta coord) coord {
	return coord{
		X: c.X + delta.X,
		Y: c.Y + delta.Y,
	}
}

type World struct {
	sync.RWMutex
	W, H     int
	entities map[coord]*entity
	players  map[string]*entity // map of player id => entity that points to that player
}

func NewWorld() *World {
	w := &World{
		W:        60,
		H:        30,
		players:  make(map[string]*entity),
		entities: make(map[coord]*entity),
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
		c := coord{3, 3}
		e = NewPlayer(playerID, c)
		w.players[playerID] = e
		w.entities[c] = e
	}
	return e
}

func (w *World) movePlayer(e *entity, a Action) {
	delta := coord{}
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

func (w *World) String() string {
	w.RLock()
	defer w.RUnlock()
	// todo use string builder (https://yourbasic.org/golang/build-append-concatenate-strings-efficiently/)
	x := 0
	y := 0
	rows := make([]string, 0)
	for y < w.H {
		row := ""
		for x < w.W {
			if e, ok := w.entities[coord{x, y}]; ok {
				row += e.String()
			} else {
				row += Space
			}
			x++
		}
		y++
		x = 0
		rows = append(rows, row)
	}
	return strings.Join(rows, "\n")
}

func (w *World) inBounds(loc coord) bool {
	return loc.X >= 0 && loc.X < w.W && loc.Y >= 0 && loc.Y < w.H
}

func (w *World) available(loc coord) bool {
	w.RLock()
	defer w.RUnlock()
	_, ok := w.entities[loc]
	return !ok
}
