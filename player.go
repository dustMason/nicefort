package main

import (
	"github.com/dustmason/nicefort/fov"
	"sync"
)

type player struct {
	sync.RWMutex
	id     string // the ssh pubkey of the connected player
	name   string
	loc    Coord
	mapMem map[Coord]string // map of what player knows on current world
	view   *fov.View

	// todo lots more stuff like inventory, stats, etc
}

func NewPlayer(id string, c Coord) *entity {
	p := &player{
		id:     id,
		loc:    c,
		mapMem: make(map[Coord]string),
		view:   fov.New(),
	}
	return &entity{class: Being, player: p}
}

func (p *player) See(w *World) {
	p.Lock()
	defer p.Unlock()
	p.view.Compute(w, p.loc.X, p.loc.Y, 10)
	for point, _ := range p.view.Visible {
		p.mapMem[Coord{point.X, point.Y}] = w.RenderForMemory(point.X, point.Y)
	}
}

func (p *player) CanSee(x, y int) bool {
	p.RLock()
	defer p.RUnlock()
	return p.view.IsVisible(x, y)
}

func (p *player) AllVisited() map[Coord]string {
	p.RLock()
	defer p.RUnlock()
	cpy := make(map[Coord]string)
	for k, v := range p.mapMem {
		cpy[k] = v
	}
	return cpy
}
