package world

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/kelindar/tile"
)

type mapView struct {
	*tile.View
}

func (pv *mapView) InBounds(x, y int) bool {
	return y > 0 && int16(y) < pv.Grid.Size.Y && x > 0 && int16(x) < pv.Grid.Size.X
}

func (pv *mapView) IsOpaque(x, y int) bool {
	return false
	// todo we need to parse and read tile data to answer this
}

func (pv *mapView) Render() string {
	pv.Each(func(pt tile.Point, t tile.Tile) {
		// tile.Point{X: int16(left), Y: int16(top)},
		// 	tile.Point{X: int16(right), Y: int16(bottom)},
		inView, dist := ply.CanSee(int(pt.X), int(pt.Y))
		memString, inPastView := seen[ic]
		if inPastView && !inView {
			b.WriteString(memString)
		} else if inView {
			// grab the tile from the grid
			var ent *entity
			ents, biome := w.allEntitiesAndBiomeAt(ix, iy)
			if len(ents) > 0 {
				fmt.Println("inview", biome, ents)
				ent = ents[0]
			} else {
				ent = &entity{environment: Grass}
			}
			// fmt.Println("render", ix, iy, ent, biome)
			bg := entityFromBiome(biome)
			b.WriteString(
				lipgloss.NewStyle().
					Foreground(lipgloss.Color(ent.ForegroundColor(dist))).
					Background(lipgloss.Color(bg.BackgroundColor(dist))).
					Render(ent.String()),
			)
		} else { // not in past or current fovView
			b.WriteString(blackSpace)
		}
		if int(pt.X) == right {
			b.WriteString("\n")
		}

	})

}
