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
	// rotating blades
	0xae: ItemInfo{
		0, 0xae, "", "object_blades", AddSpinningBlades,
	},
	// flamethrowers
	0x186: ItemInfo{
		0, 0x186, "", "object_anomaly_fire", AddFlamethrower,
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

// adds static spinning blades centered in the grid
func AddSpinningBlades(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {

	if !dusk {
		// not supported for quake
		return
	}
	entityName := item.DuskEntityName
	entity := quakemap.NewEntity(0, entityName, q)
	entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2)
	entity.OriginY = float64(y)*gridSizeY + (gridSizeY / 2)
	entity.OriginZ = gridSizeZ * 1.5
	entity.AdditionalKeys["damage"] = "10.0"
	entity.AdditionalKeys["frequency"] = "0.8"
	q.Entities = append(q.Entities, entity)
}

// adds static flamethrowers on the bottom facing up
func AddFlamethrower(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {

	if !dusk {
		// not supported for quake
		return
	}
	entityName := item.DuskEntityName
	entity := quakemap.NewEntity(0, entityName, q)
	entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2)
	entity.OriginY = float64(y)*gridSizeY + (gridSizeY / 2)
	entity.OriginZ = gridSizeZ
	q.Entities = append(q.Entities, entity)
}
