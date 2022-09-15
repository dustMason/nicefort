package main

type entity struct {
	class    Class
	subclass Subclass
	player   *player
}

type player struct {
	id   string // the ssh pubkey of the connected player
	loc  Coord
	name string

	// todo lots more stuff like inventory, stats, etc
}

func NewPlayer(id string, c Coord) *entity {
	p := &player{id: id, loc: c}
	return &entity{class: Being, player: p}
}

func (e entity) String() string {
	if e.player != nil {
		return " " + string(e.player.name[0])
	}
	return tile(e.class, e.subclass)
}

func (e entity) Color() string {
	if e.player != nil {
		return "#FDC300"
	}
	if e.subclass == Floor {
		return "#444444"
	}
	return "#fdffcc"
}

type Class int
type Subclass int

const Unknown = "? "

const (
	Being Class = iota
	Item
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
)

var tileMap = map[Class]map[Subclass]string{
	Environment: {
		WallCornerNE: "◣ ",
		WallCornerSE: "◤ ",
		WallCornerSW: "◥ ",
		WallCornerNW: "◢ ",
		Floor:        "..",
		Space:        "  ",
		Default:      "猫",
	},
	Item: {
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
