package main

import "github.com/meshiest/go-dungeon/dungeon"

func GenerateMap(size, rooms int) map[Coord]*entity {
	d := dungeon.NewDungeon(size, rooms)
	m := make(map[Coord]*entity)
	for y, row := range d.Grid {
		for x, cell := range row {
			if cell == 0 {
				c := Coord{x, y}
				m[c] = &entity{class: Environment}
			}
		}
	}
	return m
}
