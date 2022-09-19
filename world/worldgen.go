package world

import (
	"github.com/PieterD/WorldGen/worldgen2"
	"math/rand"
)

func GenerateOverworld(size int) []location {
	wg := worldgen2.NewWorld(3, 3)
	fSize := float64(size)
	m := make([]location, size*size)
	x := 0
	y := 0
	ind := 0
	for y < size {
		for x < size {
			i := wg.GetHeight_Island(float64(x)/fSize, float64(y)/fSize)
			var ent *entity
			if i < 0 {
				ent = &entity{class: Environment, subclass: Water}
			} else if i < 0.5 {
				ent = &entity{class: Environment, subclass: Grass}
			} else {
				ent = &entity{class: Environment, subclass: Rock}
			}

			// testing
			m[ind] = location{ent}
			r := rand.Intn(100)
			if r == 1 {
				m[ind] = append(m[ind], &entity{item: &TestItem, quantity: rand.Intn(4) + 1, class: Thing})
			} else if r == 2 {
				m[ind] = append(m[ind], &entity{item: &TestItem2, quantity: 1, class: Thing})
			}

			x++
			ind++
		}
		x = 0
		y++
	}

	// todo iterate the map and populate with plants and monsters

	return m
}
