package world

import (
	"github.com/lucasb-eyer/go-colorful"
	"math"
)

type entity struct {
	player      *player
	npc         *NPC
	item        *Item
	flora       *Flora
	environment Environment
	quantity    int
	variant     int // 0 < n < 10, random value decided at worldgen time to use for visual texture
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
	return environmentTile(e.environment, e.variant)
}

func (e entity) ForegroundColor(dist float64) string {
	return e.baseColor().BlendLab(dkGrey, dist).Hex()
}

func (e entity) BackgroundColor(dist float64) string {
	blend := math.Min(0.2+dist, 1.0)
	return e.baseColor().BlendLab(black, blend).Hex()
}

func (e entity) SeeThrough() bool {
	if e.flora != nil {
		return e.flora.walkable
	}
	if e.environment != None {
		return true
	}
	return e.environment != WallBlock
}

func (e entity) Memorable() bool {
	return e.npc == nil && e.player == nil
}

func (e entity) Walkable() bool {
	switch e.environment {
	case WallBlock, Water:
		return false
	}
	if e.flora != nil {
		return e.flora.walkable
	}
	return true
}

func (e entity) Pickupable() bool {
	return e.item != nil
}

func (e entity) baseColor() colorful.Color {
	if e.player != nil {
		return clr("#FDC300")
	}
	if e.flora != nil {
		return clr(e.flora.color)
	}
	switch e.environment {
	case Floor:
		return clr("#444444")
	case Water:
		h1 := "#46468C"
		h2 := "#504EA6"
		c, _ := colorful.Hex(h1)
		c2, _ := colorful.Hex(h2)
		return c.BlendLab(c2, float64(e.variant)/10.)
	case Mud:
		return clr("#3F3222")
	case Grass:
		return clr("#2B8C28")
	case Rock, Pebbles:
		return clr("#9DAAB0")
	default:
		return clr("#fdffcc")
	}
}

func clr(hex string) colorful.Color {
	c, _ := colorful.Hex(hex)
	return c
}

func (e entity) Attackable() bool {
	return e.npc != nil
}
func (e entity) Harvestable() bool {
	return e.flora != nil
}

func (e entity) Occupied() bool {
	if e.npc != nil || e.player != nil {
		return true
	}
	if e.flora != nil {
		return !e.flora.walkable
	}
	return false
}

type Environment int

const Unknown = "? "

const (
	None Environment = iota // signifies that this is not an environment
	WallBlock
	WallCornerNE
	WallCornerSE
	WallCornerSW
	WallCornerNW
	Floor
	Space
	Water
	Mud
	Grass
	Rock
	Pebbles
)

// todo ground cover to add
// Dwarf Birch
// Dwarf Juniper
// Downy Willow (#969BA4) at 200–900m on rocky mountain slopes and cliffs
// Woolly Willow (#7D8A7B) at 600-900m on rocky mountain sides
// Net-leaved Willow (#62831F) at? 300-500m on wet rocks and ledges
// Dwarf Willow (#1F3017) at 0-1500m in tundra and rock moorland

var environmentTiles = map[Environment][]string{
	WallCornerNE: {"◣ "},
	WallCornerSE: {"◤ "},
	WallCornerSW: {"◥ "},
	WallCornerNW: {"◢ "},
	Floor:        {".."},
	Space:        {"  "},
	Water:        {"≈≈"},
	Mud:          {",'", "',"},
	Grass:        {"''", "\"'"},
	Rock:         {"姅", "艫", "蠨"},
	Pebbles:      {"፨፨"},
	// Tree:         {"个", "丫"},
	// Default:      {"猫"},
}

// interesting characters
// ߷
// લ
// ૭
// ଽ
// ༒ (double?)
// ༚
// ༛
// ༜
// ტ
// ᆍ
// ࿇
// ⁂
// ⁕
// ☀ ☁
// 𓆏 lots at https://mcdlr.com/utf-8/#77641

func environmentTile(e Environment, variant int) string {
	if v, ok := environmentTiles[e]; ok {
		return v[variant%len(v)]
	}
	return Unknown
}
