package rtl

// RTL sprite info and Quake/Dusk conversion callbacks

import (
	"fmt"
	"gitlab.com/camtap/lumps/pkg/quakemap"
	"math"
)

type EntityAdderCallback func(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool)

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
	// trampolines
	0xc1: ItemInfo{
		0, 0x5a, "object_jump_pad", "object_jump_pad", AddTrampoline,
	},
}

// adds teleport entities to simulate the functionality of an elevator
func AddQuoteOnQuoteElevator(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {
}

// adds trampolines right on the floor
func AddTrampoline(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {

	if !dusk {
		// not supported for quake
		return
	}
	entity := quakemap.NewEntity(0, item.DuskEntityName, q)
	entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2)
	entity.OriginY = float64(y)*gridSizeY + (gridSizeY / 2)
	entity.OriginZ = gridSizeZ
	// could not find where "amount" was documented by NewBlood.
	// this logarithmic formula is a ballpark factor that just Seems Right(tm)
	jumpAmount := math.Log10(float64(r.FloorHeight())+0.5) * ((gridSizeZ / 64) / 2)
	entity.AdditionalKeys["amount"] = fmt.Sprintf("%02f", jumpAmount)
	q.Entities = append(q.Entities, entity)
}
