package world

import (
	"fmt"
	"github.com/PieterD/WorldGen/worldgen2"
	"math/rand"
)

func GenerateOverworld(size int) []location {
	// thresholds. each value means "up this elevation"
	water := 0.
	grassland := 0.2
	woodland := 0.3
	scrubland := 0.4
	stats := make(map[string]int)
	wg := worldgen2.NewWorld(3, 3)
	fSize := float64(size)
	m := make([]location, size*size)
	x := 0
	y := 0
	ind := 0
	for y < size {
		for x < size {
			r := rand.Intn(1000)
			v := r / 100
			i := wg.GetHeight_Island(float64(x)/fSize, float64(y)/fSize)
			var loc location
			if i < water {
				loc = location{{class: Environment, subclass: Water, variant: v}}
				stats["water"] += 1

			} else if i < grassland {
				loc = location{{class: Environment, subclass: Grass, variant: v}}
				if r == 1 {
					loc = append(loc, &entity{class: Being, npc: NewRabbit(x, y)})
				}
				stats["grass"] += 1

			} else if i < woodland {
				loc = location{{class: Environment, subclass: Grass, variant: v}}
				if r < 100 {
					loc = append(loc, &entity{class: Environment, subclass: Tree, variant: v})
				}
				stats["woods"] += 1

			} else if i < scrubland {
				loc = location{{class: Environment, subclass: Pebbles, variant: v}}
				if r == 1 {
					loc = append(loc, &entity{item: &TestItem, quantity: rand.Intn(4) + 1, class: Thing})
				} else if r == 2 {
					loc = append(loc, &entity{item: &TestItem2, quantity: 1, class: Thing})
				}
				stats["scrub"] += 1

			} else {
				loc = location{{class: Environment, subclass: Rock, variant: v}}
				stats["mountain"] += 1
			}

			if r < 10 {
				// loliphants for testing
				loc = append(loc, &entity{class: Being, npc: NewElephant(x, y)})
			}

			m[ind] = loc
			x++
			ind++
		}
		x = 0
		y++
	}
	fmt.Println("generated world", stats)
	return m
}
