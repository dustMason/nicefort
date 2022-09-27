package world

type ItemTraits int64

const (
	Weapon ItemTraits = 1 << iota
	Digger
	Axe
	Knife
	Kindling
	Fuel
	Edible
	Stick
)

type Item struct {
	ID          string
	Name        string
	Description string
	Weight      float64
	loc         Coord
	activate    func(*Item, *entity, *World) (bool, string) // accepts the player and world, returns true if the item is consumed, along with a message
	icon        string
	color       string
	traits      ItemTraits
	power       float64 // 0 < n < 1
	nonPortable bool    // when it appears, immediately drop in closest avail location. can't pick up
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

func (i Item) Activate(e *entity, w *World) (bool, string) {
	if i.activate == nil {
		return false, ""
	}
	return i.activate(&i, e, w)
}

func wieldable(i *Item, e *entity, w *World) (bool, string) {
	if e.player.wielding != nil && e.player.wielding.ID == i.ID {
		e.player.wielding = BareHands
		return false, "You put away the " + i.Name
	}
	e.player.wielding = i
	return false, "Now wielding: " + i.Name
}

type InventoryItem struct {
	Item     *Item
	Quantity int
}

func (ii InventoryItem) Weight() float64 {
	return float64(ii.Quantity) * ii.Item.Weight
}

func ActivateEdible(nutrition float64, message string) func(*Item, *entity, *World) (bool, string) {
	return func(i *Item, e *entity, w *World) (bool, string) {
		e.player.Eat(nutrition)
		return true, message
	}
}

func newItem(
	id, name, icon, color string,
	weight, power float64,
	traits ItemTraits,
	nonPortable bool,
	activate func(*Item, *entity, *World) (bool, string),
) *Item {
	return &Item{
		ID:          id,
		Name:        name,
		Weight:      weight,
		icon:        icon,
		color:       color,
		traits:      traits,
		power:       power,
		activate:    activate,
		nonPortable: nonPortable,
	}
}

var BareHands = newItem(
	"bare-hands",
	"bare hands",
	"  ",
	"#fff",
	0.0,
	0.01,
	Weapon|Digger,
	false,
	nil,
)
var SharpRock = newItem(
	"sharp-rock",
	"sharp rock",
	"r ",
	"#444",
	0.05,
	0.02,
	Knife,
	false,
	wieldable,
)
var DriedLeaves = newItem(
	"dried-leaves",
	"dried leaves",
	"d ",
	"#444",
	0.05,
	0.02,
	Kindling,
	false,
	nil,
)
var Twine = newItem(
	"twine",
	"twine",
	"t ",
	"#444",
	0.05,
	0.02,
	0,
	false,
	nil,
)
var FireStarterBow = newItem(
	"fire-starter-bow",
	"fire starter bow",
	"b ",
	"#444",
	0.05,
	0.02,
	0,
	false,
	nil,
)
var Campfire = newItem(
	"campfire",
	"Campfire",
	"f ",
	"#f00",
	0,
	0,
	0,
	true,
	nil,
)
