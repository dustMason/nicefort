package world

import (
	"encoding/binary"
	"github.com/kelindar/tile"
)

type mapTile struct {
	tile   tile.Tile
	player uint8
	biome  Biome
	npc    uint8
	item   uint16
	flora  uint16
}

func ParseTile(t tile.Tile) *mapTile {
	tt := &mapTile{tile: t}
	tt.player, tt.biome, tt.npc, tt.item, tt.flora = tileToIndexes(t)
	return tt
}

func (mt *mapTile) Tile() tile.Tile {
	var t tile.Tile
	t[0] = mt.player | (uint8(mt.biome) << 4)
	t[1] = mt.npc
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, mt.item)
	t[2] = bs[0]
	t[3] = bs[1]
	bs = make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, mt.flora)
	t[4] = bs[0]
	t[5] = bs[1]
	return t
}

// byte layout:
// 0.   [4 bits: player index][4 bits: biome index]
// 1.   npc index
// 2-3. 16 bits: item index
// 4-5. 16 bits: flora index
func tileToIndexes(t tile.Tile) (uint8, Biome, uint8, uint16, uint16) {
	player := t[0] & 0xF
	biome := t[0] >> 4
	npc := t[1]
	item := binary.LittleEndian.Uint16(t[2:4])
	flora := binary.LittleEndian.Uint16(t[4:6])
	return player, Biome(biome), npc, item, flora
}

// func indexesToTile(player, biome, npc uint8, item, flora uint16) tile.Tile {
// 	var t tile.Tile
// 	t[0] = player | (biome << 4)
// 	t[1] = npc
// 	bs := make([]byte, 2)
// 	binary.LittleEndian.PutUint16(bs, item)
// 	t[2] = bs[0]
// 	t[3] = bs[1]
// 	bs = make([]byte, 2)
// 	binary.LittleEndian.PutUint16(bs, flora)
// 	t[4] = bs[0]
// 	t[5] = bs[1]
// 	return t
// }
