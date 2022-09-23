package world

import (
	"fmt"
	"github.com/PieterD/WorldGen/worldgen2"
	"math/rand"
)

func GenerateOverworld(size int) []location {
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
	m := make([]location, size*size)

	heights := getAllHeights(size)
	for i, z := range heights {
		r := rand.Intn(1000)
		v := r / 100
		x := i % size
		y := i / size
		var loc location
		if z < water {
			loc = location{{environment: Water, variant: v}}
			stats["water"] += 1

		} else if z < bog {
			loc = location{{environment: Mud, variant: v}}
			if r < 50 {
				loc = append(loc, &entity{flora: BogMyrtle()})
			}
			stats["bog"] += 1

		} else if z < birchForest {
			loc = location{{environment: Grass, variant: v}}
			if r < 100 {
				loc = append(loc, &entity{flora: DownyBirch()})
			} else if r < 120 {
				loc = append(loc, &entity{flora: Aspen()})
			} else if r < 140 {
				loc = append(loc, &entity{flora: ScotsPine()})
			} else if r < 160 {
				loc = append(loc, &entity{flora: GreyAdler()})
			} else if r < 180 {
				loc = append(loc, &entity{flora: GoatWillow()})
			} else if r < 200 {
				loc = append(loc, &entity{flora: BirdCherry()})
			} else if r < 250 {
				// testing:
				loc = append(loc, &entity{npc: NewRabbit(x, y)})
			} else if r < 300 {
				// testing:
				loc = append(loc, &entity{npc: NewElephant(x, y)})
			}
			// todo rare arctic fox
			stats["birch forest"] += 1

		} else if z < boreal {
			loc = location{{environment: Grass, variant: v}}
			if r < 60 {
				loc = append(loc, &entity{flora: ScotsPine()})
			} else if r < 120 {
				loc = append(loc, &entity{flora: NorwaySpruce()})
			} else if r < 130 {
				loc = append(loc, &entity{flora: CloudberryBush()})
			}
			stats["boreal"] += 1

		} else if z < rocky {
			loc = location{{environment: Pebbles, variant: v}}
			if r < 60 {
				loc = append(loc, &entity{flora: ScotsPine()})
			} else if r < 120 {
				loc = append(loc, &entity{flora: NorwaySpruce()})
			}
			stats["rocky"] += 1

		} else if z < mountainous {
			loc = location{{environment: Rock, variant: v}}
			if r < 20 {
				loc = append(loc, &entity{flora: ScotsPine()})
			}
			stats["mountainous"] += 1

		} else { // glacial
			loc = location{{environment: Rock, variant: v}}
			stats["glacial"] += 1
		}

		m[i] = loc
	}

	fmt.Println("Generated world:")
	fmt.Println("counts:", stats)
	return m
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
