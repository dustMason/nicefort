package world

import (
	"fmt"
	"github.com/PieterD/WorldGen/worldgen2"
	"github.com/kelindar/tile"
	"math/rand"
)

// tile format:
// - 4b players (16 values)
// - 4b environment biome (16 values)
// - 1B npc (256 values)
// - 2B item (65,536 values) 500k of pointers
// - 2B flora (65,536 values) 500k of pointers

type Biome uint8

// maximum 16 biomes
const (
	Ocean Biome = iota
	River
	Bog
	BirchForest
	Boreal
	Rocky
	Mountainous
	Glacial
)

var allBiomes = []Biome{Ocean, River, Bog, BirchForest, Boreal, Rocky, Mountainous, Glacial}

type WorldGenerator struct {
	seed int64
}

// GenerateOverworld returns npcs, items, flora
func GenerateOverworld(grid *tile.Grid) ([]*entity, []*entity, []*entity) {
	// roughly, 1.0 == 1,000m elevation
	// thresholds below. each value means "up this elevation"
	water := 0.
	bog := 0.05
	birchForest := 0.2
	boreal := 0.4
	rocky := 0.6
	mountainous := 0.8
	// glacial := 1.0 // implied
	stats := make(map[string]int)
	npcs := make([]*entity, 0)
	items := make([]*entity, 0)
	floras := make([]*entity, 0)
	heights := getAllHeights(int(grid.Size.X))

	for i, z := range heights {
		r := rand.Intn(1000)
		var b Biome
		var npc *NPC
		var item *Item
		var flora *Flora
		if z < water {
			b = Ocean
			stats["water"] += 1

		} else if z < bog {
			b = Bog
			if r < 50 {
				flora = BogMyrtle()
				stats["flora"] += 1
			}
			stats["bog"] += 1

		} else if z < birchForest {
			b = BirchForest
			if r < 100 {
				flora = DownyBirch()
				stats["flora"] += 1
			} else if r < 120 {
				flora = Aspen()
				stats["flora"] += 1
			} else if r < 140 {
				flora = ScotsPine()
				stats["flora"] += 1
			} else if r < 160 {
				flora = GreyAdler()
				stats["flora"] += 1
			} else if r < 180 {
				flora = GoatWillow()
				stats["flora"] += 1
			} else if r < 200 {
				flora = BirdCherry()
				stats["flora"] += 1
			}
			// todo rare arctic fox
			stats["birch forest"] += 1

		} else if z < boreal {
			b = Boreal
			if r < 60 {
				flora = ScotsPine()
				stats["flora"] += 1
			} else if r < 120 {
				flora = NorwaySpruce()
				stats["flora"] += 1
			} else if r < 130 {
				flora = CloudberryBush()
				stats["flora"] += 1
			}
			stats["boreal"] += 1

		} else if z < rocky {
			b = Rocky
			if r < 60 {
				flora = ScotsPine()
				stats["flora"] += 1
			} else if r < 120 {
				flora = NorwaySpruce()
				stats["flora"] += 1
			}
			stats["rocky"] += 1

		} else if z < mountainous {
			b = Mountainous
			if r < 20 {
				flora = ScotsPine()
				stats["flora"] += 1
			}
			stats["mountainous"] += 1

		} else { // glacial
			b = Glacial
			stats["glacial"] += 1
		}

		// byte layout:
		// 1.   [4 bits: player index][4 bits: biome index]
		// 2.   npc index
		// 3-4. 16 bits: item index
		// 5-6. 16 bits: flora index

		var npcIndex uint8
		var itemIndex uint16
		var floraIndex uint16

		if npc != nil {
			npcs = append(npcs, &entity{npc: npc})
			npcIndex = uint8(len(npcs) - 1)
		}
		if item != nil {
			items = append(items, &entity{item: item, quantity: 1})
			itemIndex = uint16(len(items) - 1)
		}
		if flora != nil {
			floras = append(floras, &entity{flora: flora})
			floraIndex = uint16(len(floras) - 1)
		}
		t := indexesToTile(0, uint8(b), npcIndex, itemIndex, floraIndex)
		x := int16(i) % grid.Size.X
		y := int16(i) / grid.Size.Y
		grid.WriteAt(x, y, t)
	}

	fmt.Println("Generated world:")
	fmt.Println("counts:", stats)
	return npcs, items, floras
}

func getAllHeights(size int) []float64 {
	y := 0
	x := 0
	max := 0.
	w := worldgen2.NewWorld(3, 3)
	fSize := float64(size)
	rawHeights := make([]float64, 0)
	for y < size {
		for x < size {
			h := w.GetHeight_Island(float64(x)/fSize, float64(y)/fSize)
			rawHeights = append(rawHeights, h)
			if h > max {
				max = h
			}
			x++
		}
		x = 0
		y++
	}
	heights := make([]float64, 0, len(rawHeights))
	ratio := 1 / max
	for _, h := range rawHeights {
		var n float64
		if h > 0 {
			n = h * ratio
		} else {
			n = h
		}
		heights = append(heights, n)
	}
	return heights
}
