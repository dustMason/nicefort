package world

type entity struct {
	class    Class
	subclass Subclass
	player   *player
	npc      *NPC
	item     *Item
	quantity int
	// todo move the loc field here
}

func (e entity) String() string {
	if e.player != nil {
		return " " + string(e.player.name[0])
	}
	if e.item != nil {
		return e.item.String()
	}
	return tile(e.class, e.subclass)
}

func (e entity) Color() string {
	if e.player != nil {
		return "#FDC300"
	}
	switch e.subclass {
	case Floor:
		return "#444444"
	case Water:
		return "#315F8C"
	case Grass:
		return "#2B8C28"
	case Rock:
		return "#616267"
	default:
		return "#fdffcc"
	}
}

func (e entity) SeeThrough() bool {
	if e.class != Environment {
		return true
	}
	return e.subclass != WallBlock
}

func (e entity) Memorable() bool {
	return e.class != Being
}

func (e entity) Walkable() bool {
	if e.class != Environment {
		return false
	}
	switch e.subclass {
	case WallBlock, Water:
		return false
	default:
		return true
	}
}

func (e entity) Pickupable() bool {
	return e.item != nil
}

type Class int
type Subclass int

const Unknown = "? "

const (
	Being Class = iota
	Thing
	Environment
)

const (
	Default Subclass = iota // some classes don't need a subclass
	WallBlock
	WallCornerNE
	WallCornerSE
	WallCornerSW
	WallCornerNW
	Floor
	Space
	Water
	Grass
	Rock
)

var tileMap = map[Class]map[Subclass]string{
	Environment: {
		WallCornerNE: "◣ ",
		WallCornerSE: "◤ ",
		WallCornerSW: "◥ ",
		WallCornerNW: "◢ ",
		Floor:        "..",
		Space:        "  ",
		Water:        "≈≈",
		Grass:        "''",
		Rock:         "XX",
		Default:      "猫",
	},
	Thing: {
		Default: "i ",
	},
	Being: {
		Default: "& ",
	},
}

func tile(class Class, subclass Subclass) string {
	if submap, ok := tileMap[class]; ok {
		if m, ok := submap[subclass]; ok {
			return m
		}
		return submap[Default]
	}
	return Unknown
}
