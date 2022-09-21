package world

import "sync"

type Flora struct {
	id          string
	name        string
	icon        string
	color       string
	harvestFunc harvestFunc
}

// given an item (wielded by player) return:
// - bool: if the flora is now `dead`
// - bool: successful attempt
// - float64: amount of progress
// - []InventoryItem: items dropped
type harvestFunc func(*Item) (bool, bool, float64, []InventoryItem)

func (f Flora) String() string {
	return f.icon
}

// Harvest accepts an item wielded by the player. It has the same returns as harvestFunc
func (f *Flora) Harvest(with *Item) (bool, bool, float64, []InventoryItem) {
	return f.harvestFunc(with)
}

type product struct {
	with     ItemTraits
	depletes bool
	yields   []InventoryItem
}

func withHarvestFunc(products ...product) harvestFunc {
	state := make(map[ItemTraits]float64)
	lock := sync.Mutex{}
	return func(i *Item) (bool, bool, float64, []InventoryItem) {
		lock.Lock()
		defer lock.Unlock()
		for _, p := range products {
			if i.HasTrait(p.with) {
				if state[p.with] < 0 {
					// already exhausted this product
					return false, false, 0., nil
				}
				state[p.with] += i.Power()
				if state[p.with] < 1.0 {
					return false, true, state[p.with], nil
				}
				// if we reach here, we have just exhausted this product
				state[p.with] = -1.0
				return p.depletes, true, 1.0, p.yields
			}
		}
		return false, false, 0., nil
	}
}

func newFlora(id, name, icon, color string, h harvestFunc) *Flora {
	return &Flora{
		id:          id,
		name:        name,
		icon:        icon,
		color:       color,
		harvestFunc: h,
	}
}

func ScotsPine() *Flora {
	return newFlora(
		"scots-pine",
		"Scots Pine",
		"P ",
		"#00ff00",
		withHarvestFunc(
			product{with: Axe, depletes: true, yields: []InventoryItem{{Item: &PineWood, Quantity: 4}}},
			product{with: Knife, yields: []InventoryItem{{Item: &PineBark, Quantity: 4}}},
		),
	)
}

func NorwaySpruce() *Flora {
	return newFlora(
		"norway-spruce",
		"Norway Spruce",
		"A ",
		"#00ff00",
		withHarvestFunc(
			product{with: Axe, depletes: true, yields: []InventoryItem{{Item: &SpruceWood, Quantity: 4}}},
		),
	)
}

func Aspen() *Flora {
	return newFlora(
		"aspen",
		"Aspen",
		"AA",
		"#00ff00",
		withHarvestFunc(
			product{with: Axe, depletes: true, yields: []InventoryItem{{Item: &AspenWood, Quantity: 4}}},
		),
	)
}

// Rowan
// Grey Alder
//
// Bird Cherry
// Downy Birch
// Mountain Birch
// Dwarf Birch
// Dwarf Juniper
//
// Bog Myrtle
// Goat Willow
// Tea-leaved Willow
// Downy Willow
// Glaucous Willow
//
// Woolly Willow
// Halberd-leaved Willow
// Whortle-leaved Willow
// Net-leaved Willow
// Dwarf Willow
