package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"math"
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
	entities map[Coord]*entity
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
		w.removePlayer(playerID, e.player.loc)
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
		w.entities[c] = e
	}
	e.player.name = playerName
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
func (w *World) Render(playerID, playerName string, vw, vh int) string {
	c := w.getOrCreatePlayer(playerID, playerName).player.loc
	x := c.X
	y := c.Y

	c2, _ := colorful.Hex("#000000")

	var b strings.Builder
	b.Grow(vw*vh + vh)
	left := x - vw/2
	right := x + vw/2
	top := y - vh/2
	bottom := y + vh/2
	maxDist := float64(vw / 2)

	w.RLock()
	defer w.RUnlock()
	ix := left
	iy := top
	for iy < bottom {
		for ix < right {
			var ent *entity
			if ix < 0 || iy < 0 {
				b.WriteString(tileMap[Environment][Space])
			} else if ix > w.W-1 || iy > w.H-1 {
				b.WriteString(tileMap[Environment][Space])
			} else {
				d := dist(ix, iy, c.X, c.Y, maxDist)
				if d >= 1.0 {
					b.WriteString(tileMap[Environment][Space])
				} else {
					if e, ok := w.entities[Coord{ix, iy}]; ok {
						ent = e
					} else {
						ent = &entity{class: Environment, subclass: Floor}
					}
					c1, _ := colorful.Hex(ent.Color())
					clr := c1.BlendRgb(c2, d).Hex()
					b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(clr)).Render(ent.String()))
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

// todo this could be optimized as a lookup table of coord => distance
func dist(x1, y1, x2, y2 int, max float64) float64 {
	xd := math.Pow(float64(x2-x1), 2) * 1.25
	yd := math.Pow(float64(y2-y1), 2) * 1.25
	return math.Sqrt(xd+yd) / max
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

func (w *World) inBounds(loc Coord) bool {
	return loc.X >= 0 && loc.X < w.W && loc.Y >= 0 && loc.Y < w.H
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

func (w *World) removePlayer(playerID string, loc Coord) {
	w.Lock()
	defer w.Unlock()
	delete(w.entities, loc)
	delete(w.players, playerID)
}
