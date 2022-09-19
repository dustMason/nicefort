package world

import "math/rand"

type Item struct {
	ID          int
	Name        string
	Description string
	Weight      float64
	activate    func(*entity, *World) bool // accepts the player and world, returns true if the item is consumed
	icon        string
}

func (i Item) String() string {
	return i.icon
}

type InventoryItem struct {
	Item     *Item
	Quantity int
}

func (ii InventoryItem) Weight() float64 {
	return float64(ii.Quantity) * ii.Item.Weight
}

var TestItem = Item{
	ID:     1,
	Name:   "Test Item",
	Weight: 0.1,
	icon:   "tt",
	activate: func(e *entity, world *World) bool {
		e.player.Heal(rand.Intn(3))
		return true
	},
}

var TestItem2 = Item{
	ID:     2,
	Name:   "Test Item 2",
	Weight: 0.1,
	icon:   "22",
}

var TestItem3 = Item{
	ID:     3,
	Name:   "Test Item 3",
	Weight: 1.,
	icon:   "33",
}

var TestItem4 = Item{
	ID:     4,
	Name:   "Test Item 4",
	Weight: 1.,
	icon:   "44",
}
