package world

type ItemTraits int64

const (
	Weapon ItemTraits = 1 << iota
	Digger
	Axe
	Knife
	Edible
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

// Power is a catch-all value for the effectiveness of the item. It can signify:
// - weapon strength
// - harvesting speed
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

func ActivateEdible(nutrition float64) func(*entity, *World) bool {
	return func(e *entity, w *World) bool {
		e.player.Eat(nutrition)
		return true
	}
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
