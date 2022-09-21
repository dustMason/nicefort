package world

import (
	"github.com/lucasb-eyer/go-colorful"
	"math"
)

type entity struct {
	class    Class
	subclass Subclass
	player   *player
	npc      *NPC
	item     *Item
	flora    *Flora
	quantity int
	variant  int // 0 < n < 10, random value decided at worldgen time to use for visual texture
	// todo move the loc field here
	// todo add baseColor field and make a constructor for `entity`. baseColor should be cached
}

var black, _ = colorful.Hex("#000000")
var dkGrey, _ = colorful.Hex(memColor)

func (e entity) String() string {
	if e.player != nil {
		return " " + string(e.player.name[0])
	}
	if e.npc != nil {
		return e.npc.icon
	}
	if e.item != nil {
		return e.item.String()
	}
	if e.flora != nil {
		return e.flora.String()
	}
	return tile(e.class, e.subclass, e.variant)
}

func (e entity) ForegroundColor(dist float64) string {
	return e.baseColor().BlendLab(dkGrey, dist).Hex()
}

func (e entity) BackgroundColor(dist float64) string {
	blend := math.Min(0.2+dist, 1.0)
	return e.baseColor().BlendLab(black, blend).Hex()
}

func (e entity) SeeThrough() bool {
	if e.class != Environment {
		return true
	}
	if e.subclass == Tree {
		return false
	}
	return e.subclass != WallBlock
}

func (e entity) Memorable() bool {
	return e.class != Being
}

func (e entity) Walkable() bool {
	switch e.subclass {
	case WallBlock, Water, Tree:
		return false
	default:
		return true
	}
}

func (e entity) Pickupable() bool {
	return e.item != nil
}

func (e entity) baseColor() colorful.Color {
	var hex string
	if e.player != nil {
		hex = "#FDC300"
	}
	switch e.subclass {
	case Floor:
		hex = "#444444"
	case Water:
		h1 := "#46468C"
		h2 := "#504EA6"
		c, _ := colorful.Hex(h1)
		c2, _ := colorful.Hex(h2)
		return c.BlendLab(c2, float64(e.variant)/10.)
	case Grass, Tree:
		hex = "#2B8C28"
	case Rock, Pebbles:
		hex = "#9DAAB0"
	default:
		hex = "#fdffcc"
	}
	c, _ := colorful.Hex(hex)
	return c
}

func (e entity) Attackable() bool {
	return e.npc != nil
}
func (e entity) Harvestable() bool {
	return e.flora != nil
}

type Class int
type Subclass int

const Unknown = "? "

const (
	Being Class = iota // eg, creature/player
	Thing              // eg, item
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
	Pebbles
	Tree
)

var tileMap = map[Class]map[Subclass][]string{
	Environment: {
		WallCornerNE: {"◣ "},
		WallCornerSE: {"◤ "},
		WallCornerSW: {"◥ "},
		WallCornerNW: {"◢ "},
		Floor:        {".."},
		Space:        {"  "},
		Water:        {"≈≈"},
		Grass:        {"''", "\"'"},
		Rock:         {"姅", "艫", "蠨"},
		Pebbles:      {"፨፨"},
		Tree:         {"个", "丫"},
		Default:      {"猫"},
	},
	Thing: {
		Default: {"i "},
	},
	Being: {
		Default: {"& "},
	},
}

func tile(class Class, subclass Subclass, variant int) string {
	if submap, ok := tileMap[class]; ok {
		v := submap[Default]
		if m, ok := submap[subclass]; ok {
			v = m
		}
		return v[variant%len(v)]
	}
	return Unknown
}
