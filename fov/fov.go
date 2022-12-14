// copied from https://github.com/norendren/go-fov

package fov

import (
	"math"
)

/*
implements basic recursive shadowcasting for displaying a field of view on a 2D Grid
The exact structure of the grid has been abstracted through an interface that merely provides 3 methods
expected of any grid-based implementation
*/

// GridMap is meant to represent the basic functionality that is required to detect the opaqueness
// and boundaries of a 2D grid
type GridMap interface {
	InBounds(x, y int) bool
	IsOpaque(x, y int) bool
}

// Point to hold an X, Y position
type Point struct {
	X, Y int
}

// GridSet is maps points to distance as a percentage of radius (0 < n < 1)
type GridSet map[Point]float64

// View is the item which stores the visible set of cells any time it is called. This should be called any time
// a player's position is updated
type View struct {
	Visible GridSet
}

// New returns a new instance of an fov calculator
func New() *View {
	return &View{}
}

// Compute takes a GridMap implementation along with the X and Y coordinates representing a player's current
// position and will internally update the visibile set of tiles within the provided radius `r`
func (v *View) Compute(grid GridMap, px, py, radius int) {
	v.Visible = make(map[Point]float64)
	v.Visible[Point{px, py}] = 0.
	for i := 1; i <= 8; i++ {
		v.fov(grid, px, py, 1, 0, 1, i, radius)
	}
}

// fov does the actual work of detecting the visible tiles based on the recursive shadowcasting algorithm
// annotations provided inline below for (hopefully) easier learning
func (v *View) fov(grid GridMap, px, py, dist int, lowSlope, highSlope float64, oct, rad int) {
	// If the current distance is greater than the radius provided, then this is the end of the iteration
	if dist > rad {
		return
	}

	// Convert our slope into integers that will represent the "height" from the player position
	// "height" will alternately apply to X OR Y coordinates as we move around the octants
	low := math.Floor(lowSlope*float64(dist) + 0.5)
	high := math.Floor(highSlope*float64(dist) + 0.5)

	// inGap refers to whether we are currently scanning non-blocked tiles consecutively
	// inGap = true means that the previous tile examined was empty
	inGap := false

	for height := low; height <= high; height++ {
		// Given the player coords and a distance, height and octant, determine which tile is being visited
		mapx, mapy := distHeightXY(px, py, dist, int(height), oct)
		if grid.InBounds(mapx, mapy) && distTo(px, py, mapx, mapy) < rad {
			// As long as a tile is within the bounds of the map, if we visit it at all, it is considered visible
			// That's the efficiency of shadowcasting, you just dont visit tiles that aren't visible
			v.Visible[Point{mapx, mapy}] = float64(dist) / float64(rad)
		}

		if grid.InBounds(mapx, mapy) && grid.IsOpaque(mapx, mapy) {
			if inGap {
				// An opaque tile was discovered, so begin a recursive call
				v.fov(grid, px, py, dist+1, lowSlope, (height-0.5)/float64(dist), oct, rad)
			}
			// Any time a recursive call is made, adjust the minimum slope for all future calls within this octant
			lowSlope = (height + 0.5) / float64(dist)
			inGap = false
		} else {
			inGap = true
			// We've reached the end of the scan and, since the last tile in the scan was empty, begin
			// another on the next depth up
			if height == high {
				v.fov(grid, px, py, dist+1, lowSlope, highSlope, oct, rad)
			}
		}
	}
}

// IsVisible takes in a set of X,Y coordinates and will consult the visible set (as a gridSet) to determine
// whether that tile is visible. if visible it also returns the distance
func (v *View) IsVisible(x, y int) (bool, float64) {
	if d, ok := v.Visible[Point{x, y}]; ok {
		return true, d
	}
	return false, 0.
}

// distHeightXY performs some bitwise and operations to handle the transposition of the depth and height values
// since the concept of "depth" and "height" is relative to whichever octant is currently being scanned
func distHeightXY(px, py, d, h, oct int) (int, int) {
	if oct&0x1 > 0 {
		d = -d
	}
	if oct&0x2 > 0 {
		h = -h
	}
	if oct&0x4 > 0 {
		return px + h, py + d
	}
	return px + d, py + h
}

// distTo is simply a helper function to determine the distance between two points, for checking visibility of a tile
// within a provided radius
func distTo(x1, y1, x2, y2 int) int {
	vx := math.Pow(float64(x1-x2), 2)
	vy := math.Pow(float64(y1-y2), 2)
	return int(math.Sqrt(vx + vy))
}
