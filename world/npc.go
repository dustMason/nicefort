package world

import (
	"github.com/japanoise/dmap"
	"math/rand"
	"sync"
	"time"
)

type NPC struct {
	Name string

	sync.Mutex
	icon      string
	speed     float64 // 1.0 == every tick
	baseSpeed float64 // 1.0 == every tick
	sense     float64 // 1.0 == wakes up as soon as player sees it
	health    int
	maxHealth int
	mood      mood
	targets   map[*entity]struct{} // the player(s) to run from/to
	drop      []*InventoryItem
	behavior  behavior
	loc       Coord
	lastMoved time.Time
	dead      bool
}

type behavior func(w *World, e *entity) // todo a function that determines what this npc does next
type mood int

const (
	asleep mood = iota
	calm
	curious
	enraged
	terrorized
)

func (n *NPC) Tick(now time.Time, w *World, e *entity) {
	if now.Sub(n.lastMoved).Seconds() > (1 - n.speed) {
		n.Lock()
		defer n.Unlock()
		if !n.dead {
			n.behavior(w, e)
			n.lastMoved = now
		}
	}
}

func (n *NPC) String() string {
	return n.icon
}

func (n *NPC) Color() string {
	return "#FDC300"
}

func (n *NPC) UpdateLastMoved(t time.Time) {
	n.lastMoved = t
}

func (n *NPC) Attackable() bool {
	return true
}

// Attacked returns the amount of damage done, whether the npc is alive, and any items dropped
func (n *NPC) Attacked(e *entity, damage int) (int, bool, []*InventoryItem) {
	n.health -= damage
	n.targets[e] = struct{}{}
	if n.health <= 0 {
		n.dead = true
		return damage, true, n.drop
	}
	return damage, false, nil
}

func newNPC(name, icon string, speed float64, health int, b behavior, x, y int) *NPC {
	return &NPC{
		Name:      name,
		icon:      icon,
		baseSpeed: speed,
		health:    health,
		behavior:  b,
		loc:       Coord{x, y},
		targets:   make(map[*entity]struct{}),
	}
}

// normally doesn't care about players. runs away when attacked
func defenselessCreature(w *World, me *entity) {
	if me.npc.health < me.npc.maxHealth {
		me.npc.mood = terrorized
		me.npc.speed = me.npc.baseSpeed * 2
		// todo run away (implement a HighestNeighbor func for dmap?)
	} else {
		me.npc.mood = calm
		me.npc.speed = me.npc.baseSpeed
		x := rand.Intn(3) - 1 + me.npc.loc.X
		y := rand.Intn(3) - 1 + me.npc.loc.Y
		w.MoveNPC(x, y, me)
	}
}

// normally doesn't care about players. fights back when attacked
func annoyingCreature(w *World, me *entity) {
	if len(me.npc.targets) > 0 {
		me.npc.mood = enraged
		me.npc.speed = me.npc.baseSpeed * 3
		// todo this map should be small - only a 20x20 square around the NPC. this is slow!
		dMap := dmap.BlankDMap(w, dmap.ManhattanNeighbours)
		pt := make([]dmap.Point, 0, len(me.npc.targets))
		for t, _ := range me.npc.targets {
			pt = append(pt, t.player.loc)
		}
		dMap.Calc(pt...)
		nextMove := dMap.LowestNeighbour(me.npc.loc.X, me.npc.loc.Y)
		w.MoveNPC(nextMove.X, nextMove.Y, me)
		// todo attack the target
	} else {
		x := rand.Intn(3) - 1 + me.npc.loc.X
		y := rand.Intn(3) - 1 + me.npc.loc.Y
		w.MoveNPC(x, y, me)
	}
}

func NewRabbit(x, y int) *NPC {
	return newNPC("rabbit", "r ", 0.2, 3, defenselessCreature, x, y)
}

func NewElephant(x, y int) *NPC {
	return newNPC("elephant", "E ", 0.1, 400, annoyingCreature, x, y)
}
