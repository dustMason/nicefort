package world

import (
	"github.com/dustmason/nicefort/util"
	"github.com/japanoise/dmap"
)

// mapView is a sub-section of a larger map for the purpose of pathfinding and other
// operations too expensive to perform on the entire map
type mapView struct {
	wMap    []location
	dMap    *dmap.DijkstraMap
	targets []dmap.Point
	w       int
	h       int
	xOffset int
	yOffset int
}

func newMapView(w *World, x1, y1, x2, y2 int) mapView {
	mv := mapView{wMap: make([]location, 0)}
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
	mv.xOffset = x1 // + 1
	mv.yOffset = y1 // + 1
	dm := dmap.BlankDMap(&mv, dmap.DiagonalNeighbours)
	mv.dMap = dm
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

func (mv *mapView) String() string {
	return mv.dMap.String()
}

func (mv *mapView) index(x, y int) int {
	return y*mv.w + x
}

func (mv *mapView) calc(targets []dmap.Point) {
	adjustedOffsets := mv.adjustOffsets(targets)
	mv.dMap.Calc(adjustedOffsets...)
}

func (mv *mapView) recalc(targets []dmap.Point) {
	adjustedOffsets := mv.adjustOffsets(targets)
	mv.dMap.Recalc(adjustedOffsets...)
}

func (mv *mapView) adjustOffsets(targets []dmap.Point) []dmap.Point {
	adjustedOffsets := make([]dmap.Point, 0)
	for _, t := range targets {
		xx, yy := t.GetXY()
		ox, oy, inBounds := mv.relativePoint(xx, yy)
		if inBounds {
			adjustedOffsets = append(adjustedOffsets, Coord{X: ox, Y: oy})
		}
	}
	return adjustedOffsets
}

func (mv *mapView) lowestNeighbor(x, y int) (int, int) {
	x, y, _ = mv.relativePoint(x, y)
	wp := mv.dMap.LowestNeighbour(x, y)
	return mv.absolutePoint(wp.X, wp.Y)
}

func (mv *mapView) highestNeighbor(x, y int) (int, int) {
	x, y, _ = mv.relativePoint(x, y)
	wp := HighestNeighbor(mv.dMap, x, y)
	return mv.absolutePoint(wp.X, wp.Y)
}

// relativePoint takes a point and returns the best-fitting point within the mv.wMap
// if the point is off the map, the returned bool will be false
func (mv *mapView) relativePoint(x, y int) (int, int, bool) {
	return x - mv.xOffset, y - mv.yOffset, x-mv.xOffset > 0 && y-mv.yOffset > 0
}

func (mv *mapView) absolutePoint(x, y int) (int, int) {
	return x + mv.xOffset, y + mv.yOffset
}

func HighestNeighbor(d *dmap.DijkstraMap, x, y int) dmap.WeightedPoint {
	vals := dmap.DiagonalNeighbours(d, x, y)
	var hv dmap.Rank = 0
	ret := vals[0]
	for _, val := range vals {
		// RankMax means not passable, so those aren't candidates
		if val.Val > hv && val.Val != dmap.RankMax {
			hv = val.Val
			ret = val
		}
	}
	return ret
}
