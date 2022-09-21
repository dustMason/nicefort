package world

import "math/rand"

var lightWoodColor = "#cccc00"

type ItemTraits int64

const (
	Weapon ItemTraits = 1 << iota
	Digger
	Axe
	Knife
)

type Item struct {
	ID          string
	Name        string
	Description string
	Weight      float64
	activate    func(*entity, *World) bool // accepts the player and world, returns true if the item is consumed
	icon        string
	color       string
	traits      ItemTraits
	power       float64 // 0 < n < 1
}

func (i Item) String() string {
	return i.icon
}

func (i Item) HasTrait(t ItemTraits) bool {
	return i.traits&t == t
}

func (i Item) Damage() int {
	// todo this needs to make more sense
	// probably emit types of damage?
	return 10
}

func (i Item) Power() float64 {
	return i.power
}

type InventoryItem struct {
	Item     *Item
	Quantity int
}

func (ii InventoryItem) Weight() float64 {
	return float64(ii.Quantity) * ii.Item.Weight
}

func newItem(id, name, icon, color string, weight, power float64, traits ItemTraits, activate func(*entity, *World) bool) Item {
	return Item{
		ID:       id,
		Name:     name,
		Weight:   weight,
		icon:     icon,
		color:    color,
		traits:   traits,
		power:    power,
		activate: activate,
	}
}

var BareHands = newItem(
	"bare-hands",
	"Bare Hands",
	"  ",
	"#fff",
	0.0,
	0.01,
	Weapon|Digger,
	nil,
)

var SharpRock = newItem("sharp-rock", "Sharp Rock", "d ", "#444", 0.05, 0.02, Knife, nil)

var TestItem = newItem(
	"test-item",
	"Test Item",
	"!",
	lightWoodColor,
	0.1,
	0,
	0,
	func(e *entity, world *World) bool {
		e.player.Heal(rand.Intn(3))
		return true
	},
)

var PineWood = newItem("pine-wood", "Pine Wood", "==", lightWoodColor, 0.75, 0, 0, nil)
var PineBark = newItem("pine-bark", "Pine Bark", "~ ", lightWoodColor, 0.1, 0, 0, nil)
var SpruceWood = newItem("spruce-wood", "Spruce Wood", "==", lightWoodColor, 0.75, 0, 0, nil)
var AspenWood = newItem("aspen-wood", "Aspen Wood", "==", lightWoodColor, 0.75, 0, 0, nil)

// var TestItem2 = Item{
// 	ID:     "test-item2",
// 	Name:   "Test Item 2",
// 	Weight: 0.1,
// 	icon:   "22",
// }
//
// var TestItem3 = Item{
// 	ID:     "test-item3",
// 	Name:   "Test Item 3",
// 	Weight: 1.,
// 	icon:   "33",
// }
//
// var TestItem4 = Item{
// 	ID:     "test-item4",
// 	Name:   "Test Item 4",
// 	Weight: 1.,
// 	icon:   "44",
// }
