package main

import (
	"github.com/PieterD/WorldGen/worldgen2"
)

func GenerateOverworld(size int) map[Coord]location {
	wg := worldgen2.NewWorld(3, 3)
	fSize := float64(size)
	m := make(map[Coord]location)
	x := 0
	y := 0
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
			m[Coord{x, y}] = location{ent}
			x++
		}
		x = 0
		y++
	}

	// todo iterate the map and populate with plants and monsters

	return m
}
