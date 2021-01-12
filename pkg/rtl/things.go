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
		0, 0x2e, "weapon_nailgun", "weapon_sword", nil,
	},
	// knife
	0x2f: ItemInfo{
		0, 0x2f, "weapon_nailgun", "weapon_crossbow", nil,
	},
	// double-pistol
	0x30: ItemInfo{
		0, 0x30, "weapon_supershotgun", "weapon_pistol", nil,
	},
	// mp40
	0x31: ItemInfo{
		0, 0x31, "weapon_nailgun", "weapon_mg", nil,
	},
	// bazooka
	0x32: ItemInfo{
		0, 0x32, "weapon_rocketlauncher", "weapon_supershotgun", nil,
	},
	// firebomb
	0x33: ItemInfo{
		0, 0x33, "weapon_rocketlauncher", "weapon_riveter", nil,
	},
	// heatseaker
	0x34: ItemInfo{
		0, 0x34, "weapon_lightning", "weapon_rifle", nil,
	},
	// drunk missle
	0x35: ItemInfo{
		0, 0x35, "weapon_grenadelauncher", "weapon_mortar", nil,
	},
	// flamewall
	0x36: ItemInfo{
		0, 0x36, "weapon_lightning", "weapon_shotgun", nil,
	},
	// split missle
	0x37: ItemInfo{
		0, 0x37, "weapon_supershotgun", "weapon_supershotgun", nil,
	},
	// dark staff
	0x38: ItemInfo{
		0, 0x38, "weapon_lightning", "weapon_riveter", nil,
	},

	// powerups

	// armor
	0x10e: ItemInfo{
		0, 0x10e, "item_armor2", "item_armor2", nil,
	},

	// misc

	// trampolines
	0xc1: ItemInfo{
		0, 0x5a, "object_jump_pad", "object_jump_pad", AddTrampoline,
	},
	// rotating blades
	0xae: ItemInfo{
		0, 0xae, "", "object_blades", AddSpinningBlades,
	},
	// columns
	0xf8: ItemInfo{
		0, 0x141, "func_detail", "func_detail", AddColumn,
	},
	0xf9: ItemInfo{
		0, 0x141, "func_detail", "func_detail", AddColumn,
	},
	0xfa: ItemInfo{
		0, 0x141, "func_detail", "func_detail", AddColumn,
	},
	0xfb: ItemInfo{
		0, 0x141, "func_detail", "func_detail", AddColumn,
	},
	// push columns
	0x141: ItemInfo{
		0, 0x141, "func_train", "func_train", AddColumn,
	},
	0x165: ItemInfo{
		0, 0x141, "func_train", "func_train", AddColumn,
	},
	// flamethrowers
	0x186: ItemInfo{
		0, 0x186, "", "object_anomaly_fire", AddFlamethrower,
	},
	// fireball shooter
	0x0b: ItemInfo{
		0x0b, 0, "trap_shooter", "object_fireball_shooter", AddFireballShooter,
	},
}

// adds column or push column
func AddColumn(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {

	entity := quakemap.NewEntity(0, "func_detail", q)
	for _, brush := range quakemap.PushColumnBrushes {
		newBrush := brush.Clone()
		// NOTE: this assumes that the .map file created for it
		// has it centered at the origin
		newBrush.Scale(0.0, 0.0, 0.0, (gridSizeX / 64.0))
		entity.Brushes = append(entity.Brushes, newBrush)
	}

	entityHeight := entity.Height()
	for i, _ := range entity.Brushes {
		entity.Brushes[i].Translate(
			(float64(x)+0.5)*gridSizeX,
			(float64(y)+0.5)*-gridSizeY,
			gridSizeZ+(entityHeight/2.0))
	}

	// TODO: trigger_once for pushcolumns

	q.Entities = append(q.Entities, entity)
}

// adds trampolines right on the floor
func AddTrampoline(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {

	if !dusk {
		// not supported for quake
		return
	}
	entity := quakemap.NewEntity(0, item.DuskEntityName, q)
	entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2.0)
	entity.OriginY = float64(y)*-gridSizeY - (gridSizeY / 2.0)
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
	entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2.0)
	entity.OriginY = float64(y)*-gridSizeY - (gridSizeY / 2.0)
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
	entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2.0)
	entity.OriginY = float64(y)*-gridSizeY - (gridSizeY / 2.0)
	entity.OriginZ = gridSizeZ
	q.Entities = append(q.Entities, entity)
}

func AddFireballShooter(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {

	entityName := item.QuakeEntityName
	if dusk {
		entityName = item.DuskEntityName
	}
	actor := r.ActorGrid[y][x]

	var xoffset, yoffset, angle float64

	switch WallDirection((actor.SpriteValue - 0x8c) * 2) {
	case DIR_East:
		angle = 0.0
		xoffset = (gridSizeX / 2.0)
		yoffset = 0.0
	case DIR_North:
		angle = 90.0
		xoffset = 0.0
		yoffset = (gridSizeY / 2.0)
	case DIR_West:
		angle = 180.0
		xoffset = -(gridSizeX / 2.0)
		yoffset = 0.0
	case DIR_South:
		angle = 270.0
		xoffset = 0.0
		yoffset = -(gridSizeY / 2.0)
	}

	entity := quakemap.NewEntity(0, entityName, q)
	entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2) + xoffset
	entity.OriginY = float64(y)*-gridSizeY - (gridSizeY / 2) + yoffset
	entity.OriginZ = gridSizeZ * 1.5
	entity.AdditionalKeys["angle"] = fmt.Sprintf("%02f", angle)
	entity.AdditionalKeys["damage"] = "30"
	q.Entities = append(q.Entities, entity)
}
