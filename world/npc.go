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
	damagedBy damagedBy
	loc       Coord
	lastMoved time.Time
	dead      bool
}

type behavior func(w *World, e *entity) // todo a function that determines what this npc does next
// todo refactor damageBy to work with types of damage, not traits of items
type damagedBy func(*Item) (bool, int) // given an item ID (wielded by player) return amount of damage done
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
func (n *NPC) Attacked(by *Item, e *entity, damage int) (int, bool, bool, []*InventoryItem) {
	success, amount := n.damagedBy(by)
	n.health -= amount
	n.targets[e] = struct{}{}
	if n.health <= 0 {
		n.dead = true
		return damage, true, true, n.drop
	}
	return damage, success, false, nil
}

func newNPC(name, icon string, speed float64, health int, b behavior, x, y int) *NPC {
	return &NPC{
		Name:      name,
		icon:      icon,
		baseSpeed: speed,
		health:    health,
		behavior:  b,
		damagedBy: anyWeapon,
		loc:       Coord{x, y},
		targets:   make(map[*entity]struct{}),
	}
}

func anyWeapon(i *Item) (bool, int) {
	if i.HasTrait(Weapon) {
		return true, i.Damage()
	}
	return false, 0
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
		x1 := me.npc.loc.X - NPCActivationRadius
		y1 := me.npc.loc.Y - NPCActivationRadius
		x2 := me.npc.loc.X + NPCActivationRadius
		y2 := me.npc.loc.Y + NPCActivationRadius
		mv := newMapView(w, x1, y1, x2, y2)
		pt := make([]dmap.Point, 0, len(me.npc.targets))
		for t, _ := range me.npc.targets {
			pt = append(pt, t.player.loc)
		}
		nextX, nextY := mv.dijkstra(me.npc.loc.X, me.npc.loc.Y, pt, true)
		w.MoveNPC(nextX, nextY, me)
		// todo attack the target
	} else {
		x := rand.Intn(3) - 1 + me.npc.loc.X
		y := rand.Intn(3) - 1 + me.npc.loc.Y
		w.MoveNPC(x, y, me)
	}
}

// todo
// Elk
// Deer
// Brown Bears
// wolf
// wolverine
// lynx
// reindeer

// Salmon, trout, and the much esteemed siika (whitefish) are relatively
// abundant in the northern rivers. Baltic herring is the most common sea fish,
// while crayfish can be caught during the brief summer season. Pike, char, and
// perch are also found.

func NewRabbit(x, y int) *NPC {
	return newNPC("rabbit", "r ", 0.2, 3, defenselessCreature, x, y)
}

func NewElephant(x, y int) *NPC {
	return newNPC("elephant", "E ", 0.1, 400, annoyingCreature, x, y)
}
