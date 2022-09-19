package main

import "math/rand"

type item struct {
	id          int
	name        string
	description string
	weight      float64
	icon        string
	activate    func(*entity, *World) bool // accepts the player and world, returns true if the item is consumed
}

func (i item) String() string {
	return i.icon
}

type InventoryItem struct {
	item     *item
	quantity int
}

func (ii InventoryItem) Weight() float64 {
	return float64(ii.quantity) * ii.item.weight
}

var TestItem = item{
	id:     1,
	name:   "Test Item",
	weight: 0.1,
	icon:   "tt",
	activate: func(e *entity, world *World) bool {
		e.player.Heal(rand.Intn(3))
		return true
	},
}

var TestItem2 = item{
	id:     2,
	name:   "Test Item 2",
	weight: 0.1,
	icon:   "22",
}

var TestItem3 = item{
	id:     3,
	name:   "Test Item 3",
	weight: 1.,
	icon:   "33",
}

var TestItem4 = item{
	id:     4,
	name:   "Test Item 4",
	weight: 1.,
	icon:   "44",
}
