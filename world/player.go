package world

import (
	"github.com/dustmason/nicefort/fov"
	"math"
	"sync"
	"time"
)

type player struct {
	sync.RWMutex
	id           string // the ssh pubkey of the connected player
	name         string
	loc          Coord
	mapMem       map[Coord]string // map of what player knows on current world
	view         *fov.View
	inventoryMap map[int]*InventoryItem // map of item id => inventoryItem
	inventory    []*InventoryItem
	moveSpeed    float64 // 0 < n < 1.0

	// counters
	maxCarry  float64
	carrying  float64
	maxHealth int
	health    int
	money     int
	lastMoved time.Time // for applying moveSpeed

	// todo
	// - stats?
}

func NewPlayer(id string, c Coord) *entity {
	p := &player{
		id:           id,
		loc:          c,
		mapMem:       make(map[Coord]string),
		view:         fov.New(),
		inventoryMap: make(map[int]*InventoryItem),
		inventory:    make([]*InventoryItem, 0),
		maxCarry:     50.,
		maxHealth:    20,
		health:       20,
		money:        0,
		moveSpeed:    0.2,
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

func (p *player) PickUp(i *Item, quantity int) int {
	canCarry := math.Floor((p.maxCarry - p.carrying) / i.Weight)
	pickedUp := int(math.Min(float64(quantity), canCarry))
	if pickedUp > 0 {
		if ii, ok := p.inventoryMap[i.ID]; ok {
			ii.Quantity += pickedUp
		} else {
			nii := &InventoryItem{Item: i, Quantity: pickedUp}
			p.inventory = append(p.inventory, nii)
			p.inventoryMap[i.ID] = nii
		}
		p.carrying += float64(pickedUp) * i.Weight
	}
	// todo emit message to tell player they got a thing
	return pickedUp
}

func (p *player) Inventory() []*InventoryItem {
	return p.inventory
}

func (p *player) Heal(h int) {
	p.health += h
	if p.health > p.maxHealth {
		p.health = p.maxHealth
	}
}

func (p *player) ReplaceInventory(inv map[int]*InventoryItem) {
	p.Lock()
	defer p.Unlock()
	p.inventoryMap = inv
	ni := make([]*InventoryItem, 0)
	for _, ii := range inv {
		ni = append(ni, ii)
	}
	p.inventory = ni
}

func (p *player) SetLocation(nx, ny int, now time.Time) {
	p.lastMoved = now
	p.loc = Coord{nx, ny}
}

func (p *player) GetLocation() (int, int) {
	return p.loc.X, p.loc.Y
}

func (p *player) CanMove(now time.Time) bool {
	return now.Sub(p.lastMoved) > time.Duration(int(500.*p.moveSpeed))*time.Millisecond
}
