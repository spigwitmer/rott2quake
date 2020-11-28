package rtl

// RTL sprite info and Quake/Dusk conversion callbacks

import (
	"fmt"
	"gitlab.com/camtap/rott2quake/pkg/quakemap"
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
	// weapons

	// bat
	0x2e: ItemInfo{
		0, 0x2e, "", "weapon_sword", nil,
	},
	// knife
	0x2f: ItemInfo{
		0, 0x2f, "", "weapon_crossbow", nil,
	},
	// double-pistol
	0x30: ItemInfo{
		0, 0x30, "", "weapon_pistol", nil,
	},
	// mp40
	0x31: ItemInfo{
		0, 0x31, "", "weapon_mg", nil,
	},
	// bazooka
	0x32: ItemInfo{
		0, 0x32, "", "weapon_supershotgun", nil,
	},
	// firebomb
	0x33: ItemInfo{
		0, 0x33, "", "weapon_riveter", nil,
	},
	// heatseaker
	0x34: ItemInfo{
		0, 0x34, "", "weapon_rifle", nil,
	},
	// drunk missle
	0x35: ItemInfo{
		0, 0x35, "", "weapon_mortar", nil,
	},
	// flamewall
	0x36: ItemInfo{
		0, 0x36, "", "weapon_shotgun", nil,
	},
	// split missle
	0x37: ItemInfo{
		0, 0x37, "", "weapon_supershotgun", nil,
	},
	// dark staff
	0x38: ItemInfo{
		0, 0x38, "", "weapon_riveter", nil,
	},

	// misc

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
	// fireball shooter
	0x0b: ItemInfo{
		0x0b, 0, "trap_spikeshoote", "object_fireball_shooter", AddFireballShooter,
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

func AddFireballShooter(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {

	entityName := item.QuakeEntityName
	if dusk {
		entityName = item.DuskEntityName
	}

	// direction the fireball should go depends on a tile of id 0x0c that's aligned
	// by the x or y axis

	// traverse in each direction until the location of an 0x0c tile is
	// found. Default to west.
	angle := float64(360.0)
	var targetWallTile uint16 = 0x0c
	var curSteps int = 0
	var xoffset, yoffset float64

	// north
	if y > 0 {
		for j := y - 1; j >= 0; j-- {
			if r.ActorGrid[x][j].Type == WALL_Regular {
				if r.ActorGrid[x][j].Tile == targetWallTile {
					if y-j > curSteps {
						curSteps = y - j
						angle = 90.0
						xoffset = 0
						yoffset = -(gridSizeY / 2)
					}
				}
				break
			}
		}
	}
	// south
	if y < 127 {
		for j := y + 1; j < 128; j++ {
			if r.ActorGrid[x][j].Type == WALL_Regular {
				if r.ActorGrid[x][j].Tile == targetWallTile {
					if j-y > curSteps {
						curSteps = j - y
						angle = 270.0
						xoffset = 0
						yoffset = (gridSizeY / 2)
					}
				}
				break
			}
		}
	}
	// west
	if x > 0 {
		for i := x - 1; i >= 0; i-- {
			if r.ActorGrid[i][y].Type == WALL_Regular {
				if r.ActorGrid[i][y].Tile == targetWallTile {
					if x-i > curSteps {
						curSteps = x - i
						angle = 0.0
						xoffset = -(gridSizeX / 2)
						yoffset = 0
					}
				}
				break
			}
		}
	}
	// east
	if x < 127 {
		for i := x + 1; i < 128; i++ {
			if r.ActorGrid[i][y].Type == WALL_Regular {
				if r.ActorGrid[i][y].Tile == targetWallTile {
					if i-x > curSteps {
						angle = 180.0
						xoffset = (gridSizeX / 2)
						yoffset = 0
					}
				}
				break
			}
		}
	}

	entity := quakemap.NewEntity(0, entityName, q)
	entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2) + xoffset
	entity.OriginY = float64(y)*gridSizeY + (gridSizeY / 2) + yoffset
	entity.OriginZ = gridSizeZ * 1.5
	entity.AdditionalKeys["angle"] = fmt.Sprintf("%02f", angle)
	q.Entities = append(q.Entities, entity)
}
