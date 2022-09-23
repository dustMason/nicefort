package world

import (
	"github.com/kelindar/tile"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTile(t *testing.T) {
	rawTile := tile.Tile{0, 0, 0, 0, 0, 0}
	mt := ParseTile(rawTile)
	assert.Equal(t, 0, mt.player)
	assert.Equal(t, 0, mt.biome)
	assert.Equal(t, 0, mt.npc)
	assert.Equal(t, 0, mt.item)
	assert.Equal(t, 0, mt.flora)
}
