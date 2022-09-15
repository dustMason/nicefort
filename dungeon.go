package main

import "github.com/meshiest/go-dungeon/dungeon"

func GenerateMap(size, rooms int) map[Coord]*entity {
	d := dungeon.NewDungeon(size, rooms)
	m := make(map[Coord]*entity)

	for y, row := range d.Grid {
		for x, cell := range row {
			if cell == 0 {
				// e := smoothTile(d.Grid, x, y)
				e := &entity{class: Environment, subclass: WallBlock}
				c := Coord{x, y}
				m[c] = e
			}
		}
	}
	return m
}

func smoothTile(grid [][]int, x, y int) *entity {
	score := 0

	// if the tile above is a wall, add 1
	if isSet(grid, x, y-1) {
		score += 1
	}

	// if the tile to the right is a wall, add 2
	if isSet(grid, x+1, y) {
		score += 2
	}

	// if the tile below is a wall, add 4
	if isSet(grid, x, y+1) {
		score += 4
	}

	// if the tile to the left is a wall, add 8
	if isSet(grid, x-1, y) {
		score += 8
	}

	e := &entity{class: Environment}

	switch score {
	case 3:
		e.subclass = WallCornerSW
	case 6:
		e.subclass = WallCornerNW
	case 9:
		e.subclass = WallCornerSE
	case 12:
		e.subclass = WallCornerNE
	default:
		e.subclass = WallBlock
	}

	return e
}

func isSet(grid [][]int, x, y int) bool {
	if y < 0 || y > len(grid)-1 {
		return false
	}
	row := grid[y]
	if x < 0 || x > len(row)-1 {
		return false
	}
	return row[x] == 0
}
