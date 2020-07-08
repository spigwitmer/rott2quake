package rtl

// RTL sprite info and Quake/Dusk conversion callbacks

import (
	"gitlab.com/camtap/lumps/pkg/quakemap"
)

type EntityAdderCallback func(x int, y int, w *WallInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool)

type ItemInfo struct {
	TileId          uint16              // is it represented by a tile (can be 0)
	SpriteId        uint16              // is it represented by a sprite (can be 0)
	QuakeEntityName string              // replacement Quake entity name
	DuskEntityName  string              // replacement Dusk entity name
	AddCallback     EntityAdderCallback // callback function (takes precedence over replacement entity names)
}

var Items = map[uint16]ItemInfo{
	// use teleporters to implement elevators
	0x5a: ItemInfo{
		0, 0x5a, "", "", AddQuoteOnQuoteElevator,
	},
}

// adds teleport entities to simulate the functionality of an elevator
func AddQuoteOnQuoteElevator(x int, y int, w *WallInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {
}
