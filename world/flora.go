package world

import "sync"

type Flora struct {
	id          string
	name        string
	icon        string
	color       string
	harvestFunc harvestFunc
	walkable    bool
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

func newFlora(id, name, icon, color string, walkable bool, h harvestFunc) *Flora {
	return &Flora{
		id:          id,
		name:        name,
		icon:        icon,
		color:       color,
		harvestFunc: h,
		walkable:    walkable,
	}
}

func newFloraProduct(id, name, icon, color string, weight float64, activate func(*entity, *World) bool) *Item {
	return &Item{ID: id, Name: name, Weight: weight, icon: icon, color: color, traits: 0, power: 0, activate: activate}
}

// ScotsPine
// from sea level to 1,000m
// The species is mainly found on poorer, sandy soils, rocky outcrops, peat bogs or close to the forest limit.
//
// In Scandinavian countries, the pine was used for making tar in the preindustrial age
// The wood is pale brown to red-brown, and used for general construction work
// The pine fibres are used to make the textile known as vegetable flannel
//
//	The raw fibre, called Waldwolle ("forest wool"), and the pine oil were separated, and then the Waldwolle was spun into yarn or thread, and either woven or knitted.
func ScotsPine() *Flora {
	return newFlora(
		"scots-pine",
		"Scots Pine",
		"P ",
		"#3F3C18",
		false,
		withHarvestFunc(
			product{with: Axe, depletes: true, yields: []InventoryItem{{Item: PineWood, Quantity: 4}}},
			product{with: Knife, yields: []InventoryItem{{Item: PineBark, Quantity: 4}}},
		),
	)
}

var PineBark = newFloraProduct("pine-bark", "Pine Bark", "~ ", "#372F22", 0.1, nil)
var PineWood = newFloraProduct("pine-wood", "Pine Wood", "==", "#EE852D", 0.75, nil)

// NorwaySpruce
// The tree is the source of spruce beer, which was once used to prevent and even cure scurvy.
// Norway spruce shoot tips have been used in traditional Austrian medicine internally (as syrup or tea)
// and externally (as baths, for inhalation, as ointments, as resin application or as tea) for treatment
// of disorders of the respiratory tract, skin, locomotor system, gastrointestinal tract and infections.
func NorwaySpruce() *Flora {
	return newFlora(
		"norway-spruce",
		"Norway Spruce",
		"A ",
		"#424118",
		false,
		withHarvestFunc(
			product{with: Axe, depletes: true, yields: []InventoryItem{{Item: SpruceWood, Quantity: 4}}},
			product{with: Knife, yields: []InventoryItem{{Item: SpruceShoots, Quantity: 1}}},
		),
	)
}

var SpruceWood = newFloraProduct("spruce-wood", "Spruce Wood", "==", "#C0A18C", 0.75, nil)
var SpruceShoots = newFloraProduct("spruce-shoots", "Spruce Shoots", "u ", "#5B7B1A", 0.01, nil)

// It is also a popular animal bedding
// 0-100m?
func Aspen() *Flora {
	return newFlora(
		"aspen",
		"Aspen",
		"AA",
		"#388164",
		false,
		withHarvestFunc(
			product{with: Axe, depletes: true, yields: []InventoryItem{{Item: AspenWood, Quantity: 4}}},
			product{with: Knife, yields: []InventoryItem{{Item: AspenBark, Quantity: 4}}},
		),
	)
}

var AspenWood = newFloraProduct("aspen-wood", "Aspen Wood", "==", "#E9D6AC", 0.75, nil)
var AspenBark = newFloraProduct("aspen-bark", "Aspen Bark", "~ ", "#848582", 0.1, nil)

// GreyAdler
// The Ho-Chunk people eat the bark of the rugosa subspecies when their stomachs are "sour" or upset.
// In the northern part of its range, it is a common tree species at sea level in forests, abandoned fields and on lakeshore
// A. rugosa provides cover for wildlife, is browsed by deer and moose, and the seeds are eaten by birds
func GreyAdler() *Flora {
	return newFlora(
		"grey-adler",
		"Grey Adler",
		"A ",
		"#78A14D",
		false,
		withHarvestFunc(
			product{with: Axe, depletes: true, yields: []InventoryItem{{Item: GreyAdlerWood, Quantity: 4}}},
			product{with: Knife, yields: []InventoryItem{{Item: GreyAdlerBark, Quantity: 2}}},
		),
	)
}

var GreyAdlerWood = newFloraProduct("grey-adler-wood", "Grey Adler Wood", "==", "#AB8458", 0.75, nil)
var GreyAdlerBark = newFloraProduct("grey-adler-bark", "Grey Adler Bark", "~ ", "#BF8C77", 0.75, nil)

func BirdCherry() *Flora {
	return newFlora(
		"bird-cherry",
		"Bird Cherry",
		"A ",
		"#3C840B",
		false,
		withHarvestFunc(
			product{with: Axe, depletes: true, yields: []InventoryItem{{Item: BirdCherryWood, Quantity: 4}}},
			product{with: 0, yields: []InventoryItem{{Item: BirdCherries, Quantity: 4}}},
		),
	)
}

var BirdCherryWood = newFloraProduct("bird-cherry-wood", "Bird Cherry Wood", "==", "#A77235", 0.75, nil)
var BirdCherries = newFloraProduct("bird-cherries", "Bird Cherries", "፝፝", "#0C1A1B", 0.05, nil)

// DownyBirch
// The outer layer of bark can be stripped off the tree without killing it and can be used to make canoe skins, drinking vessels and roofing tiles.
// The inner bark can be used for the production of rope and for making a form of oiled paper.
// This bark is also rich in tannin and has been used as a brown dye and as a preservative.
// The bark can also be turned into a high quality charcoal favoured by artists.
// The twigs and young branches are very flexible and make good whisks and brooms.
// The timber is pale in colour with a fine, uniform texture and is used in the manufacture of plywood, furniture, shelves, coffins, matches and toys, and in turnery.
func DownyBirch() *Flora {
	return newFlora(
		"downy-birch",
		"Downy Birch",
		"A ",
		"#876E3A",
		false,
		withHarvestFunc(
			product{with: Axe, depletes: true, yields: []InventoryItem{{Item: DownyBirchWood, Quantity: 4}, {Item: DownyBirchBranches, Quantity: 4}}},
			product{with: Knife, yields: []InventoryItem{{Item: DownyBirchBark, Quantity: 4}, {Item: DownyBirchBranches, Quantity: 2}}},
		),
	)
}

var DownyBirchWood = newFloraProduct("downy-birch-wood", "Downy Birch Wood", "==", "#BB926B", 0.75, nil)
var DownyBirchBark = newFloraProduct("downy-birch-bark", "Downy Birch Bark", "~ ", "#8C827E", 0.1, nil)
var DownyBirchBranches = newFloraProduct("downy-birch-branches", "Downy Birch Branches", "~~", "#676153", 0.25, nil)

// BogMyrtle
// The foliage has a sweet resinous scent and is a traditional insect repellent, used by campers to keep biting insects out of tents.
// It is also a traditional component of royal wedding bouquets and is used variously in perfumery and as a condiment.
// The leaves can be dried to make tea, and both the nutlets and dried leaves can be used to make a seasoning.
// In some native cultures in Eastern Canada, the plant has been used as a traditional remedy for stomach aches, fever, bronchial ailments, and liver problems
func BogMyrtle() *Flora {
	return newFlora(
		"bog-myrtle",
		"Bog Myrtle",
		"m ",
		"#6C8568",
		true,
		withHarvestFunc(
			product{with: 0, yields: []InventoryItem{{Item: BogMyrtleLeaves, Quantity: 4}}},
		),
	)
}

var BogMyrtleLeaves = newFloraProduct("bog-myrtle-leaves", "Bog Myrtle Leaves", "..", "#778872", 0.01, nil)

// GoatWillow
// In Scandinavia it has been fairly common to make willow flutes from goat willow cuttings.
func GoatWillow() *Flora {
	return newFlora(
		"goat-willow",
		"Goat Willow",
		"w ",
		"#B7C052",
		true,
		withHarvestFunc(
			product{with: Knife, yields: []InventoryItem{{Item: GoatWillowStalks, Quantity: 2}}},
		),
	)
}

var GoatWillowStalks = newFloraProduct("goat-willow-stalks", "Goat Willow Stalks", "..", "#4D5824", 0.05, nil)

// GlaucousWillow
// was used by Native Americans as a painkiller
func GlaucousWillow() *Flora {
	return newFlora(
		"glaucous-willow",
		"Glaucous Willow",
		"W ",
		"#233812",
		true,
		withHarvestFunc(
			product{with: 0, yields: []InventoryItem{{Item: GlaucousWillowCatkins, Quantity: 2}}},
		),
	)
}

var GlaucousWillowCatkins = newFloraProduct("glaucous-willow-catkins", "Glaucous Willow Catkins", ",,", "#CFBDA8", 0.01, nil)

// HalberdLeavedWillow
// Native Americans used parts of willows, including this species, for medicinal purposes, basket weaving, to make bows and arrows, and for building animal traps.
// In Yukon, willow leaves were chewed to treat mosquito bites and bee stings, as well as stomach aches
func HalberdLeavedWillow() *Flora {
	return newFlora(
		"halberd-leaved-willow",
		"Halberd-leaved Willow",
		"w ",
		"#EAE29D",
		true,
		withHarvestFunc(
			product{with: Knife, yields: []InventoryItem{{Item: HalberdLeavedWillowSticks, Quantity: 3}}},
			product{with: 0, yields: []InventoryItem{{Item: HalberdLeavedWillowLeaves, Quantity: 4}}},
		),
	)
}

var HalberdLeavedWillowSticks = newFloraProduct("halberd-leaved-willow-sticks", "Halberd Leaved Willow Sticks", "--", "#7E9D0B", 0.1, nil)
var HalberdLeavedWillowLeaves = newFloraProduct("halberd-leaved-willow-leaves", "Halberd Leaved Willow Leaves", "o ", "#7E9D0B", 0.01, nil)

// CloudberryBush
func CloudberryBush() *Flora {
	return newFlora(
		"cloudberry-bush",
		"Cloudberry Bush",
		"w ",
		"#C34105",
		true,
		withHarvestFunc(
			product{with: 0, yields: []InventoryItem{{Item: Cloudberries, Quantity: 10}}},
		),
	)
}

var Cloudberries = newFloraProduct("cloudberries", "Cloudberries", ". ", "#FAB3BD", 0.01, ActivateEdible(0.05))
