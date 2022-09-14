package main

type entity struct {
	class  Class
	player *player
}

type player struct {
	id     string
	secret string
	loc    Coord
	// todo lots more stuff like inventory, stats, etc
}

func NewPlayer(id string, c Coord) *entity {
	p := &player{id: id, loc: c}
	return &entity{class: Being, player: p}
}

func (e entity) String() string {
	switch e.class {
	case Being:
		if e.player != nil {
			return "J" // todo get char from player
		}
		return "&"
	case Item:
		return "i"
	case Environment:
		return "%"
	case Portal:
		return "d"
	case Projectile:
		return "-" // todo bearing and velocity
	default:
		return ""
	}
}

type Class int

const (
	Being Class = iota
	Item
	Environment
	Portal
	Projectile
)
