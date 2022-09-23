package world

import (
	"github.com/dustmason/nicefort/util"
	"github.com/japanoise/dmap"
)

// mapView is a sub-section of a larger map for the purpose of pathfinding and other
// operations too expensive to perform on the entire map
type mapView struct {
	wMap    []location
	targets []dmap.Point
	w       int
	h       int
	xOffset int
	yOffset int
}

func newMapView(w *World, targets []dmap.Point, x1, y1, x2, y2 int) mapView {
	mv := mapView{wMap: make([]location, 0), targets: targets}
	y1 = util.ClampedInt(y1, 0, w.H)
	x1 = util.ClampedInt(x1, 0, w.W)
	iy := y1
	for iy < y2 {
		rowStart := iy*w.W + x1
		rowEnd := iy*w.W + x2
		mv.wMap = append(mv.wMap, w.wMap[rowStart:rowEnd]...)
		iy++
		if iy >= w.H {
			break
		}
	}
	mv.w = x2 - x1
	mv.h = iy - y1
	mv.xOffset = x1 + 1
	mv.yOffset = y1 + 1
	return mv
}

// SizeX SizeY IsPassable and OOB to satisfy the dmap interface
func (mv *mapView) SizeX() int {
	return mv.w
}

func (mv *mapView) SizeY() int {
	return mv.h
}

func (mv *mapView) IsPassable(x int, y int) bool {
	for _, e := range mv.wMap[mv.index(x, y)] {
		if !e.Walkable() {
			return false
		}
	}
	return true
}

func (mv *mapView) OOB(x int, y int) bool {
	return !(x >= 0 && x < mv.w && y >= 0 && y < mv.h)
}

func (mv *mapView) index(x, y int) int {
	return y*mv.w + x
}

// todo implement moveTowards = false, ie a HighestNeighbor func on dmap
func (mv *mapView) dijkstra(x, y int, moveTowards bool) (int, int) {
	// all given points must be relative to the smaller slice of the map, and then translated back
	x, y, _ = mv.offsetPoint(x, y)
	adjustedOffsets := make([]dmap.Point, 0)
	for _, t := range mv.targets {
		xx, yy := t.GetXY()
		ox, oy, inBounds := mv.offsetPoint(xx, yy)
		if inBounds {
			adjustedOffsets = append(adjustedOffsets, Coord{X: ox, Y: oy})
		}
	}
	dMap := dmap.BlankDMap(mv, dmap.DiagonalNeighbours)
	dMap.Calc(adjustedOffsets...)
	var nextMove dmap.WeightedPoint
	if moveTowards {
		nextMove = dMap.LowestNeighbour(x, y)
	} else {
		nextMove = HighestNeighbor(dMap, x, y)
	}
	return nextMove.X + mv.xOffset, nextMove.Y + mv.yOffset
}

// offsetPoint takes a point and returns the best-fitting point within the mv.wMap
// if the point is off the map, the returned bool will be false
func (mv *mapView) offsetPoint(x, y int) (int, int, bool) {
	return x - mv.xOffset, y - mv.yOffset, x-mv.xOffset > 0 && y-mv.yOffset > 0
}

// todo this doesn't seem to work. NPCs just get stuck in a cycle
func HighestNeighbor(d *dmap.DijkstraMap, x, y int) dmap.WeightedPoint {
	vals := dmap.DiagonalNeighbours(d, x, y)
	var hv dmap.Rank = 0
	ret := vals[0]
	for _, val := range vals {
		if val.Val > hv {
			hv = val.Val
			ret = val
		}
	}
	return ret
}
