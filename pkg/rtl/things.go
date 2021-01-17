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
	TileId          uint16 // is it represented by a tile (can be 0)
	SpriteId        uint16 // is it represented by a sprite (can be 0)
	QuakeEntityName string // replacement Quake entity name
	DuskEntityName  string // replacement Dusk entity name
	QuakeHeight     float64
	DuskHeight      float64
	PlaceOnFloor    bool
	AddCallback     EntityAdderCallback // callback function (takes precedence over replacement entity names)
}

var Items = map[uint16]ItemInfo{
	// weapons

	// bat
	0x2e: ItemInfo{
		0, 0x2e, "weapon_nailgun", "weapon_sword", 0, 0, false, nil,
	},
	// knife
	0x2f: ItemInfo{
		0, 0x2f, "weapon_nailgun", "weapon_crossbow", 0, 0, false, nil,
	},
	// double-pistol
	0x30: ItemInfo{
		0, 0x30, "weapon_supershotgun", "weapon_pistol", 0, 0, false, nil,
	},
	// mp40
	0x31: ItemInfo{
		0, 0x31, "weapon_nailgun", "weapon_mg", 0, 0, false, nil,
	},
	// bazooka
	0x32: ItemInfo{
		0, 0x32, "weapon_rocketlauncher", "weapon_supershotgun", 0, 0, false, nil,
	},
	// firebomb
	0x33: ItemInfo{
		0, 0x33, "weapon_rocketlauncher", "weapon_riveter", 0, 0, false, nil,
	},
	// heatseaker
	0x34: ItemInfo{
		0, 0x34, "weapon_lightning", "weapon_rifle", 0, 0, false, nil,
	},
	// drunk missle
	0x35: ItemInfo{
		0, 0x35, "weapon_grenadelauncher", "weapon_mortar", 0, 0, false, nil,
	},
	// flamewall
	0x36: ItemInfo{
		0, 0x36, "weapon_lightning", "weapon_shotgun", 0, 0, false, nil,
	},
	// split missle
	0x37: ItemInfo{
		0, 0x37, "weapon_supershotgun", "weapon_supershotgun", 0, 0, false, nil,
	},
	// dark staff
	0x38: ItemInfo{
		0, 0x38, "weapon_lightning", "weapon_riveter", 0, 0, false, nil,
	},

	// powerups

	// armor
	0x10e: ItemInfo{
		0, 0x10e, "item_armor2", "item_armor2", 0, 0, false, nil,
	},

	// misc

	// trampolines
	0xc1: ItemInfo{
		0, 0x5a, "object_jump_pad", "object_jump_pad", 0, 0, false, AddTrampoline,
	},
	// rotating blades
	0xae: ItemInfo{
		0, 0xae, "", "object_blades", 0, 0, false, AddSpinningBlades,
	},
	// columns
	0xf8: ItemInfo{
		0, 0x141, "func_detail", "func_detail", 0, 0, false, AddColumn,
	},
	0xf9: ItemInfo{
		0, 0x141, "func_detail", "func_detail", 0, 0, false, AddColumn,
	},
	0xfa: ItemInfo{
		0, 0x141, "func_detail", "func_detail", 0, 0, false, AddColumn,
	},
	0xfb: ItemInfo{
		0, 0x141, "func_detail", "func_detail", 0, 0, false, AddColumn,
	},
	// push columns
	0x141: ItemInfo{
		0, 0x141, "func_train", "func_train", 0, 0, false, AddColumn,
	},
	0x165: ItemInfo{
		0, 0x141, "func_train", "func_train", 0, 0, false, AddColumn,
	},
	// exploding barrels
	0x10d: ItemInfo{
		0, 0x10d, "misc_explobox", "misc_explobox", 64, 64, true, nil,
	},
	0x3e: ItemInfo{
		0, 0x10d, "misc_explobox", "misc_explobox", 64, 64, true, nil,
	},
	// exploding box
	0x3d: ItemInfo{
		0, 0x10d, "misc_explobox2", "misc_explobox2", 32, 32, true, nil,
	},
	// flamethrowers
	0x186: ItemInfo{
		0, 0x186, "", "object_anomaly_fire", 0, 0, false, AddFlamethrower,
	},
	// fireball shooter
	0x0b: ItemInfo{
		0x0b, 0, "trap_shooter", "object_fireball_shooter", 0, 0, false, AddFireballShooter,
	},
}

// adds column or push column
func AddColumn(x int, y int, gridSizeX float64, gridSizeY float64, gridSizeZ float64,
	item *ItemInfo, r *RTLMapData, q *quakemap.QuakeMap, dusk bool) {

	actor := &r.ActorGrid[y][x]
	entityType := "func_detail"

	var touchx1, touchy1, touchx2, touchy2 float64
	touchdx := 0
	touchdy := 0
	pushTrigger := false

	touchz1 := gridSizeZ
	touchz2 := gridSizeZ * 2

	switch actor.SpriteValue {
	// rt_ted.c:3805
	case 303, 304, 305, // pushes east
		321, 322, 323, // pushes north
		339, 340, 341, // pushes west
		357, 358, 359: // pushes south
		entityType = "func_train"
		pushTrigger = true
	}

	entity := quakemap.NewEntity(0, entityType, q)
	AddDefaultEntityKeys(entity, actor)
	for _, brush := range quakemap.PushColumnBrushes {
		newBrush := brush.Clone()
		// NOTE: this assumes that the .map file created for it
		// has it centered at the origin
		newBrush.Scale(0.0, 0.0, 0.0, (gridSizeX / 64.0))
		entity.Brushes = append(entity.Brushes, newBrush)
	}

	// build trigger_once right at the edge of the column
	switch actor.SpriteValue {
	case 303, 304, 305:
		touchx1 = float64(x)*gridSizeX + (gridSizeX-entity.Width())/2.0
		touchx2 = touchx1 + 1
		touchy1 = float64(y) * -gridSizeY
		touchy2 = touchy1 - gridSizeY
		touchdx = 1
	case 321, 322, 323:
		touchy1 = float64(y)*-gridSizeY - (gridSizeY+entity.Length())/2.0
		touchy2 = touchy1 - 1
		touchx1 = float64(x) * gridSizeX
		touchx2 = touchx1 + gridSizeX
		touchdy = -1
	case 339, 340, 341:
		touchx1 = float64(x)*gridSizeX + (gridSizeX+entity.Width())/2.0
		touchx2 = touchx1 + 1
		touchy1 = float64(y) * -gridSizeY
		touchy2 = touchy1 - gridSizeY
		touchdx = -1
	case 357, 358, 359:
		touchy1 = float64(y)*-gridSizeY - (gridSizeY-entity.Length())/2.0
		touchy2 = touchy1 - 1
		touchx1 = float64(x) * gridSizeX
		touchx2 = touchx1 + gridSizeX
		touchdy = 1
	}

	entityHeight := entity.Height()
	for i, _ := range entity.Brushes {
		entity.Brushes[i].Translate(
			(float64(x)+0.5)*gridSizeX,
			(float64(y)+0.5)*-gridSizeY,
			gridSizeZ+(entityHeight/2.0))
	}

	if pushTrigger {
		var touchplateX, touchplateY int
		tgtColumn := fmt.Sprintf("column_%d_%d", x, y)
		tgtPathStart := fmt.Sprintf("column_%d_%d_corner_start", x, y)
		tgtPathEnd := fmt.Sprintf("column_%d_%d_corner_end", x, y)
		tgtRelay := fmt.Sprintf("trigger_%d_%d", x, y)
		if actor.InfoValue > 0 {
			touchplateX = int((actor.InfoValue >> 8) & 0xff)
			touchplateY = int(actor.InfoValue & 0xff)
			tgtRelay = fmt.Sprintf("trigger_%d_%d", touchplateX, touchplateY)
		}

		entity.AdditionalKeys["targetname"] = tgtColumn
		entity.AdditionalKeys["target"] = tgtPathStart

		initialCorner := quakemap.NewEntity(0, "path_corner", q)
		initialCorner.OriginX = float64(x)*gridSizeX + (gridSizeX-entity.Width())/2.0
		initialCorner.OriginY = float64(y+1)*-gridSizeY + (gridSizeY-entity.Length())/2.0
		initialCorner.OriginZ = gridSizeZ
		initialCorner.AdditionalKeys["targetname"] = tgtPathStart
		initialCorner.AdditionalKeys["target"] = tgtPathEnd
		q.Entities = append(q.Entities, initialCorner)

		// only move 1 unit instead of 2 if there's a wall
		if !r.ActorGrid[y+(touchdy*2)][x+(touchdx*2)].IsWall() {
			touchdx *= 2
			touchdy *= 2
		}

		endCorner := quakemap.NewEntity(0, "path_corner", q)
		endCorner.OriginX = initialCorner.OriginX + (float64(touchdx) * gridSizeX)
		endCorner.OriginY = initialCorner.OriginY - (float64(touchdy) * gridSizeY)
		endCorner.OriginZ = gridSizeZ
		endCorner.AdditionalKeys["targetname"] = tgtPathEnd
		endCorner.AdditionalKeys["target"] = "idontexist"
		endCorner.AdditionalKeys["wait"] = "-1"
		q.Entities = append(q.Entities, endCorner)

		pushEntityRelay := quakemap.NewEntity(0, "trigger_relay", q)
		pushEntityRelay.OriginX = (float64(x) + 0.5) * gridSizeZ
		pushEntityRelay.OriginY = (float64(y) + 0.5) * -gridSizeY
		pushEntityRelay.OriginZ = gridSizeZ * 2.5
		pushEntityRelay.AdditionalKeys["targetname"] = tgtRelay
		pushEntityRelay.AdditionalKeys["target"] = tgtColumn
		q.Entities = append(q.Entities, pushEntityRelay)

		if actor.InfoValue == 0 {
			pushEntity := quakemap.NewEntity(0, "trigger_once", q)
			pushEntity.Brushes = append(pushEntity.Brushes, quakemap.BasicCuboid(
				touchx1, touchy1, touchz1,
				touchx2, touchy2, touchz2,
				"clip", (gridSizeX/64.0), false,
			))
			pushEntity.AdditionalKeys["target"] = tgtRelay
			q.Entities = append(q.Entities, pushEntity)
		} else {
			r.AddTrigger(actor, touchplateX, touchplateY, TRIGGER_WallPush)
		}

	}

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
