package world

import (
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
	target    *entity // the player to run from/to
	drop      []*InventoryItem
	behavior  behavior
	loc       Coord
	lastMoved time.Time
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
		n.behavior(w, e)
		n.lastMoved = now
	}
}

func (n *NPC) Attacked(e *entity, damage int) []*InventoryItem {
	n.health -= damage
	n.target = e
	if n.health <= 0 {
		return n.drop
	}
	return nil
}

// used by animals who normally don't care about players. they move aimlessly until attacked, and
// then try to run away.
func defenselessCreature(w *World, me *entity) {
	if me.npc.health < me.npc.maxHealth {
		me.npc.mood = terrorized
		me.npc.speed = me.npc.baseSpeed * 2
		// todo run away
	} else {
		me.npc.mood = calm
		me.npc.speed = me.npc.baseSpeed
		dx := rand.Intn(3) - 1
		dy := rand.Intn(3) - 1
		w.MoveNPC(dx, dy, me)
	}
}

func NewRabbit(x, y int) *NPC {
	return &NPC{
		Name:      "rabbit",
		icon:      "of",
		baseSpeed: 0.2,
		health:    3,
		behavior:  defenselessCreature,
		loc:       Coord{x, y},
	}
}
