package rtl

import (
	"fmt"
	"gitlab.com/camtap/rott2quake/pkg/quakemap"
	"log"
)

// RTL to Quake MAP conversion functions
//
// Keep in mind that in ROTT, Y increases southward while in
// Trenchbroom, Y increases northward, so Y axis values are
// inverted.

var (
	exitLumps = []string{
		"EXIT",
		"ENTRANCE",
		"EXITARCH",
		"EXITARCA",
		"ENTRARCH",
	}
)

func ClassNameForMaskedWall(w *MaskedWallInfo, position string) string {
	if w.IsSwitch && position == "above" {
		return "func_button"
	}
	passable := false  // can we walk through without any other action?
	shootable := false // can we pass through it after we shoot it?
	if position == "above" && w.Flags&MWF_AbovePassable > 0 {
		passable = true
	}
	if position == "bottom" && w.Flags&MWF_BottomPassable > 0 {
		passable = true
	}
	if position == "middle" && w.Flags&MWF_MiddlePassable > 0 {
		passable = true
	}
	if position == "bottom" && w.Flags&(MWF_Shootable|MWF_BlockingChanges) == (MWF_Shootable|MWF_BlockingChanges) {
		shootable = true
	}
	if shootable {
		return "func_breakable"
	}
	if passable {
		return "func_illusionary"
	}
	return "func_detail"
}

type ElevatorNode struct {
	Switch ActorInfo
	Floor  ActorInfo
}

// Adds func_button and trigger_teleport entities to link elevators
func LinkElevators(rtlmap *RTLMapData, textureWad string,
	floorDepth, gridSizeX, gridSizeY, gridSizeZ, scale float64,
	dusk bool, qm *quakemap.QuakeMap) {
	elevators := make(map[uint16][]ElevatorNode)

	elevatorSwitchTile := uint16(0x4c)
	elevatorDoorTile := uint16(0x66)

	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			// look for elevator door, calculate floor from where switch
			// is
			if rtlmap.WallPlane[y][x] == elevatorDoorTile {
				if x > 1 && rtlmap.WallPlane[y][x-2] == elevatorSwitchTile {
					linkCode := rtlmap.SpritePlane[y][x-1]
					elevators[linkCode] = append(elevators[linkCode], ElevatorNode{
						Floor:  rtlmap.ActorGrid[y][x-1],
						Switch: rtlmap.ActorGrid[y][x-2],
					})
				} else if x < 126 && rtlmap.WallPlane[y][x+2] == elevatorSwitchTile {
					linkCode := rtlmap.SpritePlane[y][x+1]
					elevators[linkCode] = append(elevators[linkCode], ElevatorNode{
						Floor:  rtlmap.ActorGrid[y][x+1],
						Switch: rtlmap.ActorGrid[y][x+2],
					})
				} else if y > 1 && rtlmap.WallPlane[y-2][x] == elevatorSwitchTile {
					linkCode := rtlmap.SpritePlane[y-1][x]
					elevators[linkCode] = append(elevators[linkCode], ElevatorNode{
						Floor:  rtlmap.ActorGrid[y-1][x],
						Switch: rtlmap.ActorGrid[y-2][x],
					})
				} else if y < 126 && rtlmap.WallPlane[y+2][x] == elevatorSwitchTile {
					linkCode := rtlmap.SpritePlane[y+1][x]
					elevators[linkCode] = append(elevators[linkCode], ElevatorNode{
						Floor:  rtlmap.ActorGrid[y+1][x],
						Switch: rtlmap.ActorGrid[y+2][x],
					})
				}
			}
		}
	}

	for linkCode, elevatorNodes := range elevators {
		if len(elevatorNodes) != 2 {
			log.Printf("Elevator tiles with link code %d != 2???", len(elevatorNodes))
			continue
		}

		elev1 := elevatorNodes[0]
		elev2 := elevatorNodes[1]

		var floor1ButtonX1, floor1ButtonY1, floor1ButtonX2, floor1ButtonY2 float64
		var floor2ButtonX1, floor2ButtonY1, floor2ButtonX2, floor2ButtonY2 float64
		var button1Angle, button2Angle string

		if elev1.Switch.X != elev1.Floor.X {
			// func_button facing east or west
			floor1ButtonX1 = float64(elev1.Floor.X) * gridSizeX
			floor1ButtonY1 = float64(elev1.Floor.Y) * -gridSizeY
			floor1ButtonY2 = floor1ButtonY1 - gridSizeY
			if elev1.Switch.X > elev1.Floor.X {
				floor1ButtonX1 += gridSizeX
				button1Angle = "180"
			} else {
				button1Angle = "0"
			}
			floor1ButtonX2 = floor1ButtonX1 + 1
		} else {
			// func_button facing north or south
			floor1ButtonY1 = float64(elev1.Floor.Y) * -gridSizeY
			floor1ButtonX1 = float64(elev1.Floor.X) * gridSizeX
			floor1ButtonX2 = floor1ButtonX1 + gridSizeX
			if elev1.Switch.Y > elev1.Floor.Y {
				floor1ButtonY1 -= gridSizeY
				button1Angle = "90"
			} else {
				button1Angle = "270"
			}
			floor1ButtonY2 = floor1ButtonY1 - 1
		}
		if elev2.Switch.X != elev2.Floor.X {
			floor2ButtonX1 = float64(elev2.Floor.X) * gridSizeX
			floor2ButtonY1 = float64(elev2.Floor.Y) * -gridSizeY
			floor2ButtonY2 = floor2ButtonY1 - gridSizeY
			if elev2.Switch.X > elev2.Floor.X {
				floor2ButtonX1 += gridSizeX
				button2Angle = "180"
			} else {
				button2Angle = "0"
			}
			floor2ButtonX2 = floor2ButtonX1 + 1
		} else {
			floor2ButtonY1 = float64(elev2.Floor.Y) * -gridSizeY
			floor2ButtonX1 = float64(elev2.Floor.X) * gridSizeX
			floor2ButtonX2 = floor2ButtonX1 + gridSizeX
			if elev2.Switch.Y > elev2.Floor.Y {
				floor2ButtonY1 -= gridSizeY
				button2Angle = "90"
			} else {
				button2Angle = "270"
			}
			floor2ButtonY2 = floor2ButtonY1 - 1
		}

		// elevator 1
		floor1Entity := quakemap.NewEntity(0, "func_button", qm)
		floor1Entity.AdditionalKeys["target"] = fmt.Sprintf("elev_%d_1_trigger", linkCode)
		floor1Entity.AdditionalKeys["angle"] = button1Angle
		floor1Entity.AdditionalKeys["lip"] = "1"
		floor1Brush := quakemap.BasicCuboid(floor1ButtonX1, floor1ButtonY1, floorDepth,
			floor1ButtonX2, floor1ButtonY2, float64(rtlmap.FloorHeight()+1)*gridSizeZ,
			"ELEV5", scale, false)
		floor1Entity.Brushes = append(floor1Entity.Brushes, floor1Brush)
		qm.Entities = append(qm.Entities, floor1Entity)

		floor1TriggerEntity := quakemap.NewEntity(0, "trigger_teleport", qm)
		floor1TriggerEntity.AdditionalKeys["targetname"] = fmt.Sprintf("elev_%d_1_trigger", linkCode)
		floor1TriggerEntity.AdditionalKeys["target"] = fmt.Sprintf("elev_%d_1", linkCode)
		floor1TriggerEntityBrush := quakemap.BasicCuboid(float64(elev1.Floor.X)*gridSizeX, float64(elev1.Floor.Y)*-gridSizeY, floorDepth,
			float64(elev1.Floor.X+1)*gridSizeX, float64(elev1.Floor.Y+1)*-gridSizeY, floorDepth+gridSizeZ,
			"__TB_empty", -1.0, false)
		floor1TriggerEntity.Brushes = append(floor1TriggerEntity.Brushes, floor1TriggerEntityBrush)
		qm.Entities = append(qm.Entities, floor1TriggerEntity)

		floor1DestEntity := quakemap.NewEntity(0, "info_teleport_destination", qm)
		floor1DestEntity.OriginX = float64(elev2.Floor.X)*gridSizeX + (gridSizeX / 2)
		floor1DestEntity.OriginY = float64(elev2.Floor.Y)*-gridSizeY - (gridSizeY / 2)
		floor1DestEntity.OriginZ = floorDepth
		floor1DestEntity.AdditionalKeys["targetname"] = fmt.Sprintf("elev_%d_1", linkCode)
		floor1DestEntity.AdditionalKeys["angle"] = button2Angle
		qm.Entities = append(qm.Entities, floor1DestEntity)

		// elevator 2
		floor2Entity := quakemap.NewEntity(0, "func_button", qm)
		floor2Entity.AdditionalKeys["target"] = fmt.Sprintf("elev_%d_2_trigger", linkCode)
		floor2Entity.AdditionalKeys["angle"] = button2Angle
		floor2Entity.AdditionalKeys["lip"] = "1"
		floor2Brush := quakemap.BasicCuboid(floor2ButtonX1, floor2ButtonY1, floorDepth,
			floor2ButtonX2, floor2ButtonY2, float64(rtlmap.FloorHeight()+1)*gridSizeZ,
			"ELEV5", scale, false)
		floor2Entity.Brushes = append(floor2Entity.Brushes, floor2Brush)
		qm.Entities = append(qm.Entities, floor2Entity)

		floor2TriggerEntity := quakemap.NewEntity(0, "trigger_teleport", qm)
		floor2TriggerEntity.AdditionalKeys["targetname"] = fmt.Sprintf("elev_%d_2_trigger", linkCode)
		floor2TriggerEntity.AdditionalKeys["target"] = fmt.Sprintf("elev_%d_2", linkCode)
		floor2TriggerEntityBrush := quakemap.BasicCuboid(float64(elev2.Floor.X)*gridSizeX, float64(elev2.Floor.Y)*-gridSizeY, floorDepth,
			float64(elev2.Floor.X+1)*gridSizeX, float64(elev2.Floor.Y+1)*-gridSizeY, floorDepth+gridSizeZ,
			"__TB_empty", -1.0, false)
		floor2TriggerEntity.Brushes = append(floor2TriggerEntity.Brushes, floor2TriggerEntityBrush)
		qm.Entities = append(qm.Entities, floor2TriggerEntity)

		floor2DestEntity := quakemap.NewEntity(0, "info_teleport_destination", qm)
		floor2DestEntity.OriginX = float64(elev1.Floor.X)*gridSizeX + (gridSizeX / 2)
		floor2DestEntity.OriginY = float64(elev1.Floor.Y)*-gridSizeY - (gridSizeY / 2)
		floor2DestEntity.OriginZ = floorDepth
		floor2DestEntity.AdditionalKeys["targetname"] = fmt.Sprintf("elev_%d_2", linkCode)
		floor2DestEntity.AdditionalKeys["angle"] = button1Angle
		qm.Entities = append(qm.Entities, floor2DestEntity)
	}
}

func CreateThinWall(rtlmap *RTLMapData, x, y int, scale float64, qm *quakemap.QuakeMap) {
	var x1, y1, x2, y2 float64
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	infoVal := rtlmap.InfoPlane[y][x]
	actor := rtlmap.ActorGrid[y][x]
	texName := actor.WallTileToTextureName(false)

	if infoVal == 1 || (infoVal >= 4 && infoVal <= 9) {
		if actor.ThinWallDirection == WALLDIR_NorthSouth {
			x1 = float64(x)*gridSizeX + (gridSizeX / 2) - 2
			x2 = float64(x)*gridSizeX + (gridSizeX / 2) + 2
			y1 = float64(y) * -gridSizeY
			y2 = float64(y+1) * -gridSizeY
		} else {
			x1 = float64(x) * gridSizeX
			x2 = float64(x+1) * gridSizeX
			y1 = float64(y)*-gridSizeY - ((gridSizeY / 2.0) - 2.0)
			y2 = float64(y)*-gridSizeY - ((gridSizeY / 2.0) + 2.0)
		}
		switch infoVal {
		case 1:
			// above passable
			var z1 float64 = floorDepth
			var z2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
		case 4:
			// above only
			var z1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			var z2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
		case 5:
			// below only
			var z1 float64 = floorDepth
			var z2 float64 = floorDepth + gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
		case 6:
			// middle passable
			var bottomz1 float64 = floorDepth
			var bottomz2 float64 = floorDepth + gridSizeZ
			var topz1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			var topz2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
			wallColumn1 := quakemap.BasicCuboid(x1, y1, bottomz1, x2, y2, bottomz2,
				texName, scale, false)
			wallColumn2 := quakemap.BasicCuboid(x1, y1, topz1, x2, y2, topz2,
				texName, scale, false)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn1)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn2)
		case 7:
			// everything but below
			var z1 float64 = floorDepth + gridSizeZ
			var z2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
		case 8:
			// middle only
			var z1 float64 = floorDepth + gridSizeZ
			var z2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
		case 9:
			// everything but above
			var z1 float64 = floorDepth
			var z2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
		}
	}
}

func CreateTrigger(rtlmap *RTLMapData, actor *ActorInfo, scale float64, qm *quakemap.QuakeMap) {
	switch actor.Type {
	case WALL_MaskedWall:
		// rendered in CreateMaskedWall
	case WALL_Regular:
		CreateWallSwitchTrigger(rtlmap, actor, scale, qm)
	default:
		CreateTouchplate(rtlmap, actor, scale, qm)
	}
}

func CreateWallSwitchTrigger(rtlmap *RTLMapData, actor *ActorInfo, scale float64, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	x1 := float64(actor.X) * gridSizeX
	y1 := float64(actor.Y) * -gridSizeY
	z1 := floorDepth
	x2 := float64(actor.X+1) * gridSizeX
	y2 := float64(actor.Y+1) * -gridSizeY
	z2 := floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ

	// build column that overlaps the wall
	wallColumnBrush := quakemap.BasicCuboid(x1, y1, z1,
		x2, y2, z2,
		"__TB_empty", scale, true)
	triggerEntity := quakemap.NewEntity(0, "trigger_multiple", qm)
	triggerEntity.AdditionalKeys["target"] = fmt.Sprintf("trigger_%d_%d_relay", actor.X, actor.Y)
	triggerEntity.AdditionalKeys["message"] = "Switch Triggered."
	triggerEntity.Brushes = append(triggerEntity.Brushes, wallColumnBrush)
	qm.Entities = append(qm.Entities, triggerEntity)
}

func CreateTouchplate(rtlmap *RTLMapData, actor *ActorInfo, scale float64, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	relayTargetName := fmt.Sprintf("trigger_%d_%d_relay", actor.X, actor.Y)

	triggerEntity := quakemap.NewEntity(0, "trigger_once", qm)
	triggerEntity.AdditionalKeys["target"] = relayTargetName
	triggerEntity.AdditionalKeys["message"] = "Touchplate Triggered"
	triggerEntity.Brushes = append(triggerEntity.Brushes,
		quakemap.BasicCuboid(float64(actor.X)*gridSizeX, float64(actor.Y)*-gridSizeY, floorDepth,
			float64(actor.X+1)*gridSizeX, float64(actor.Y+1)*-gridSizeY, floorDepth+gridSizeZ,
			"__TB_empty", scale, false))
	qm.Entities = append(qm.Entities, triggerEntity)
}

func CreateRegularWall(rtlmap *RTLMapData, x, y int, scale float64, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	infoVal := rtlmap.InfoPlane[y][x]
	spriteVal := rtlmap.SpritePlane[y][x]
	actor := rtlmap.ActorGrid[y][x]
	texName := actor.WallTileToTextureName(false)
	var initialCorner *quakemap.Entity
	var moveWallInfo MoveWallInfo

	entityType := "func_wall"

	if actor.Tile == 0x4c {
		// do not render elevator switches as those get spawned as
		// buttons later
		return
	}

	x1 := float64(x) * gridSizeX
	y1 := float64(y) * -gridSizeY
	z1 := floorDepth
	x2 := float64(x+1) * gridSizeX
	y2 := float64(y+1) * -gridSizeY
	z2 := floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ

	// plain ol' column
	wallColumn := quakemap.BasicCuboid(x1, y1, z1,
		x2, y2, z2,
		texName, scale, true)

	if actor.MapFlags&WALLFLAGS_Moving != 0 {
		// implement moving walls as func_trains
		// TODO: switchplate activated walls
		// TODO: push activated walls

		var lastPathCorner, currentPathCorner *quakemap.Entity
		pathType, wallPath, numNodes := rtlmap.DetermineWallPath(&actor, (infoVal > 0 || spriteVal < 256))
		moveWallInfo = MoveWallSpriteIDs[spriteVal]
		initialCorner = quakemap.NewEntity(0, "path_corner", qm)
		cornerZ := floorDepth
		initialCorner.OriginX = (float64(x)) * gridSizeX
		initialCorner.OriginY = (float64(y) + 1) * -gridSizeY
		initialCorner.OriginZ = cornerZ
		initialCorner.AdditionalKeys["targetname"] = fmt.Sprintf("movewallpath_%d_%d_init", actor.X, actor.Y)
		lastPathCorner = initialCorner
		qm.Entities = append(qm.Entities, lastPathCorner)

		currentNode := wallPath
		nodeToTargetNames := make(map[*PathNode]string)
		for i := 0; i < numNodes; i++ {
			currentPathCorner = quakemap.NewEntity(0, "path_corner", qm)
			currentPathCorner.OriginZ = cornerZ
			currentPathCorner.OriginX = (float64(currentNode.X)) * gridSizeX
			currentPathCorner.OriginY = (float64(currentNode.Y) + 1) * -gridSizeY
			targetName := fmt.Sprintf("movewallpath_%d_%d_%d", actor.X, actor.Y, i)
			nodeToTargetNames[currentNode] = targetName
			currentPathCorner.AdditionalKeys["targetname"] = targetName
			lastPathCorner.AdditionalKeys["target"] = targetName
			qm.Entities = append(qm.Entities, currentPathCorner)
			currentNode = currentNode.Next
			lastPathCorner = currentPathCorner
		}

		if pathType == PATH_Perpetual {
			currentPathCorner.AdditionalKeys["target"] = nodeToTargetNames[currentNode]
		} else {
			lastPathCorner.AdditionalKeys["target"] = "idontexist"
			lastPathCorner.AdditionalKeys["wait"] = "-1"
		}

		entityType = "func_train"
	}

	// make static walls part of the worldspawn,
	// everything else a separate entity
	if spriteVal == 0 && infoVal == 0 && !actor.Damage {
		qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
	} else {
		entity := quakemap.NewEntity(0, entityType, qm)
		entity.Brushes = []quakemap.Brush{wallColumn}
		entity.AdditionalKeys["_x"] = fmt.Sprintf("%d", actor.X)
		entity.AdditionalKeys["_y"] = fmt.Sprintf("%d", actor.Y)
		if initialCorner != nil {
			entity.OriginX = initialCorner.OriginX
			entity.OriginY = initialCorner.OriginY
			entity.OriginZ = initialCorner.OriginZ
			entity.AdditionalKeys["target"] = initialCorner.AdditionalKeys["targetname"]

			wallTargetName := fmt.Sprintf("movewallpath_%d_%d_wall", actor.X, actor.Y)

			if infoVal > 0 {
				entity.AdditionalKeys["targetname"] = wallTargetName

				triggerX := int(infoVal>>8) & 0xff
				triggerY := int(infoVal) & 0xff

				// create trigger_relay to match the touchplate/switch
				relayEntity := quakemap.NewEntity(0, "trigger_relay", qm)
				relayEntity.OriginX = (float64(actor.X) + 0.5) * gridSizeX
				relayEntity.OriginY = (float64(actor.Y) + 0.5) * -gridSizeY
				relayEntity.OriginZ = floorDepth + (float64(rtlmap.FloorHeight()+1))*gridSizeZ
				relayEntity.AdditionalKeys["targetname"] = fmt.Sprintf("trigger_%d_%d_relay", triggerX, triggerY)
				relayEntity.AdditionalKeys["target"] = wallTargetName
				qm.Entities = append(qm.Entities, relayEntity)
			} else if spriteVal < 256 {
				// add pushwall trigger_once entity within the wall
				pushWallTriggerEntity := quakemap.NewEntity(0, "trigger_once", qm)
				pushWallTriggerEntity.Brushes = append(pushWallTriggerEntity.Brushes,
					quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2, "__TB_empty", scale, true))
				pushWallTriggerEntity.AdditionalKeys["_x"] = fmt.Sprintf("%d", actor.X)
				pushWallTriggerEntity.AdditionalKeys["_y"] = fmt.Sprintf("%d", actor.Y)
				pushWallTriggerEntity.AdditionalKeys["message"] = "Push Wall Activated."
				pushWallTriggerEntity.AdditionalKeys["target"] = wallTargetName
				pushWallTriggerEntity.AdditionalKeys["targetname"] = fmt.Sprintf("movewallpath_%d_%d_push", actor.X, actor.Y)
				qm.Entities = append(qm.Entities, pushWallTriggerEntity)
				entity.AdditionalKeys["targetname"] = fmt.Sprintf("movewallpath_%d_%d_wall", actor.X, actor.Y)
			}
			entity.AdditionalKeys["speed"] = fmt.Sprintf("%d", moveWallInfo.Speed*64)
		}
		qm.Entities = append(qm.Entities, entity)

		if entityType == "func_wall" && actor.Damage {
			// add trigger_hurt over wall to mimic a damaging wall
			hurtEntity := quakemap.NewEntity(0, "trigger_hurt", qm)
			hurtEntity.Brushes = append(hurtEntity.Brushes, quakemap.BasicCuboid(x1, y1, z1,
				x2, y2, z2,
				texName, scale, true))
			hurtEntity.AdditionalKeys["_x"] = fmt.Sprintf("%d", actor.X)
			hurtEntity.AdditionalKeys["_y"] = fmt.Sprintf("%d", actor.Y)
			qm.Entities = append(qm.Entities, hurtEntity)
		}
	}
}

func CreatePlatform(rtlmap *RTLMapData, x, y int, scale float64, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	actor := rtlmap.ActorGrid[y][x]
	floorHeight := rtlmap.FloorHeight()

	// platforms are supposed to work like masked walls,
	// however just implement that as full walls since
	// they're made to allow the player walk across
	if platformInfo, ok := Platforms[actor.PlatformID]; ok {
		x1 := float64(x)*gridSizeX + 1
		y1 := float64(y)*-gridSizeY - 1
		x2 := float64(x+1)*gridSizeX - 1
		y2 := float64(y+1)*-gridSizeY + 1

		// above as separate entity
		// NOTE: don't render tops and bottoms of platforms
		// if they're passable, they look nasty
		if platformInfo.Above != "" && platformInfo.Flags&MWF_AbovePassable == 0 && floorHeight > 1 {
			var abovez1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			var abovez2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
			aboveClassName := "func_detail"
			aboveColumn := quakemap.BasicCuboid(x1, y1, abovez1, x2, y2, abovez2,
				"{"+platformInfo.Above,
				scale, false)
			aboveEntity := quakemap.NewEntity(0, aboveClassName, qm)
			aboveEntity.Brushes = append(aboveEntity.Brushes, aboveColumn)
			qm.Entities = append(qm.Entities, aboveEntity)
		}

		// middle
		if platformInfo.Middle != "" && floorHeight > 2 {
			var middlez1 float64 = floorDepth + gridSizeZ
			var middlez2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			mwColumn := quakemap.BasicCuboid(x1, y1, middlez1, x2, y2, middlez2,
				"{"+platformInfo.Middle,
				scale, false)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, mwColumn)
		}

		// bottom
		// above as separate entity
		if platformInfo.Bottom != "" && platformInfo.Flags&MWF_BottomPassable == 0 {
			var z1 float64 = floorDepth + 1
			var z2 float64 = floorDepth + gridSizeZ
			className := ClassNameForMaskedWall(&platformInfo, "bottom")
			column := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				"{"+platformInfo.Bottom,
				scale, false)
			bottomEntity := quakemap.NewEntity(0, className, qm)
			bottomEntity.Brushes = append(bottomEntity.Brushes, column)
			qm.Entities = append(qm.Entities, bottomEntity)
		}
	}
}

func CreateMaskedWall(rtlmap *RTLMapData, x, y int, scale float64, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	wallInfo := rtlmap.ActorGrid[y][x]
	floorHeight := rtlmap.FloorHeight()

	// masked walls have adjacent sides, a thin wall in the
	// middle, and the bottom may be passable
	if maskedWallInfo, ok := MaskedWalls[wallInfo.Tile]; ok {
		wallDirection, _, _ := rtlmap.ThinWallDirection(x, y)
		var x1, y1, x2, y2 float64

		if wallDirection == WALLDIR_NorthSouth {
			x1 = float64(x)*gridSizeX + (gridSizeX / 2)
			x2 = float64(x)*gridSizeX + (gridSizeX / 2) + 1
			y1 = float64(y) * -gridSizeY
			y2 = float64(y+1) * -gridSizeY
		} else {
			x1 = float64(x) * gridSizeX
			x2 = float64(x+1) * gridSizeX
			y1 = float64(y)*-gridSizeY - (gridSizeY / 2.0)
			y2 = float64(y)*-gridSizeY - (gridSizeY / 2.0) + 1
		}

		// above as separate entity
		if maskedWallInfo.Above != "" && floorHeight > 1 {
			var abovez1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			var abovez2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
			aboveClassName := ClassNameForMaskedWall(&maskedWallInfo, "above")
			aboveColumn := quakemap.BasicCuboid(x1, y1, abovez1, x2, y2, abovez2,
				"{"+maskedWallInfo.Above,
				scale, false)
			aboveEntity := quakemap.NewEntity(0, aboveClassName, qm)
			aboveEntity.Brushes = append(aboveEntity.Brushes, aboveColumn)
			if maskedWallInfo.IsSwitch {
				aboveEntity.AdditionalKeys["target"] = fmt.Sprintf("trigger_%d_%d", wallInfo.X, wallInfo.Y)
				aboveEntity.AdditionalKeys["lip"] = fmt.Sprintf("%.02f", 64.0*scale)
			}
			qm.Entities = append(qm.Entities, aboveEntity)
		}

		// middle
		if maskedWallInfo.Middle != "" && floorHeight > 2 {
			var middlez1 float64 = floorDepth + gridSizeZ
			var middlez2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			middleClassName := ClassNameForMaskedWall(&maskedWallInfo, "middle")
			mwColumn := quakemap.BasicCuboid(x1, y1, middlez1, x2, y2, middlez2,
				"{"+maskedWallInfo.Middle,
				scale, false)
			middleEntity := quakemap.NewEntity(0, middleClassName, qm)
			middleEntity.Brushes = append(middleEntity.Brushes, mwColumn)
			qm.Entities = append(qm.Entities, middleEntity)
		}

		// bottom
		// above as separate entity
		if maskedWallInfo.Bottom != "" {
			var z1 float64 = floorDepth
			var z2 float64 = floorDepth + gridSizeZ
			className := ClassNameForMaskedWall(&maskedWallInfo, "bottom")
			column := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				"{"+maskedWallInfo.Bottom,
				scale, false)
			bottomEntity := quakemap.NewEntity(0, className, qm)
			bottomEntity.Brushes = append(bottomEntity.Brushes, column)
			qm.Entities = append(qm.Entities, bottomEntity)
		}

		// TODO: sides

	} else {
		panic(fmt.Sprintf("Masked wall at %d,%d has non-existent ID (%d)", x, y, wallInfo.MaskedWallID))
	}
}

func CreateDoorEntities(rtlmap *RTLMapData, scale float64, dusk bool, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	// determine which keys to use
	whichKey := 0
	availKeys := []string{"item_key1", "item_key2"}
	if dusk {
		availKeys = []string{"key_red_key", "key_blue_key", "key_yellow_key"}
	}
	keyMap := make(map[DoorLock]string)

	var timedTriggerEntity *quakemap.Entity

	for doornum, door := range rtlmap.GetDoors() {
		doorEntity := quakemap.NewEntity(0, "func_door", qm)
		timeBeforeOpen := 0
		for _, doorTile := range door.Tiles {
			if doorTile.Type != WALL_Door {
				panic(fmt.Sprintf("(%d,%d) not WALL_Door type!", doorTile.X, doorTile.Y))
			}
			texInfo := GetDoorTextures(doorTile.Tile)
			if doorTile.InfoValue > 0 {
				timeBeforeOpen = int(doorTile.InfoValue>>8) * 60
			}
			var x1, y1, x2, y2, abovex1, abovey1, abovex2, abovey2 float64
			var z1 float64 = floorDepth
			var z2 float64 = floorDepth + gridSizeZ
			if door.Direction == WALLDIR_NorthSouth {
				x1 = float64(doorTile.X)*gridSizeX + (gridSizeX / 2)
				x2 = float64(doorTile.X)*gridSizeX + (gridSizeX / 2) + 1
				y1 = float64(doorTile.Y) * -gridSizeY
				y2 = float64(doorTile.Y+1) * -gridSizeY
				// give the wall above the door 1 more pixel
				// or else it looks weird when opened
				abovey1 = y1
				abovey2 = y2
				abovex1 = x1 - 1
				abovex2 = x2 + 1
			} else {
				x1 = float64(doorTile.X) * gridSizeX
				x2 = float64(doorTile.X+1) * gridSizeX
				y1 = float64(doorTile.Y)*-gridSizeY - (gridSizeY / 2.0)
				y2 = float64(doorTile.Y)*-gridSizeY - ((gridSizeY / 2.0) + 1)
				abovey1 = y1 + 1
				abovey2 = y2 - 1
				abovex1 = x1
				abovex2 = x2
			}
			doorBrush := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texInfo.BaseTexture,
				scale, false)
			doorEntity.Brushes = append(doorEntity.Brushes, doorBrush)
			aboveBrush := quakemap.BasicCuboid(abovex1, abovey1, z2,
				abovex2, abovey2, floorDepth+float64(rtlmap.FloorHeight())*gridSizeZ,
				texInfo.AltTexture,
				scale, false)
			qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, aboveBrush)
		}

		doorEntity.AdditionalKeys["grid_start_x"] = fmt.Sprintf("%d", door.Tiles[0].X)
		doorEntity.AdditionalKeys["grid_start_y"] = fmt.Sprintf("%d", door.Tiles[0].Y)

		if door.Lock != LOCK_Unlocked && door.Lock != LOCK_Trigger {
			if dusk {
				doorEntity.AdditionalKeys["key"] = fmt.Sprintf("%d", whichKey+1)
			} else {
				doorEntity.AdditionalKeys["spawnflags"] = fmt.Sprintf("%d", whichKey+1)
			}

			if _, ok := keyMap[door.Lock]; !ok {
				// place keys on the map
				for y := 0; y < 128; y++ {
					for x := 0; x < 128; x++ {
						if rtlmap.ActorGrid[y][x].Type == ACTOR_None && rtlmap.SpritePlane[y][x] == uint16(door.Lock+0x1c) {
							entity := quakemap.NewEntity(0, availKeys[whichKey], qm)
							entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2)
							entity.OriginY = float64(y)*-gridSizeY - (gridSizeY / 2.0)
							entity.OriginZ = floorDepth + (gridSizeZ / 2)
							if dusk {
								// FIXME when/if dusk SDK fixes keys
								// being placed a lot lower than the
								// entity's origin
								entity.OriginZ += 600.0
							}
							qm.Entities = append(qm.Entities, entity)
						}
					}
				}

				keyMap[door.Lock] = availKeys[whichKey]
				whichKey = (whichKey + 1) % len(availKeys)
				if whichKey == 0 {
					log.Printf("More than %d keys used, this map may not be playable (or fun)", len(availKeys))
				}
			}

		}

		doorEntity.AdditionalKeys["_doornum"] = fmt.Sprintf("%d", doornum)
		// move upward when open
		if door.Tiles[0].Tile == 0x66 { // but move elevator door sideways
			elevTileX, elevTileY := door.Tiles[0].X, door.Tiles[0].Y
			if elevTileX < 126 && rtlmap.WallPlane[elevTileY][elevTileX+2] == 0x4c {
				doorEntity.AdditionalKeys["angle"] = "90"
			} else if elevTileX > 1 && rtlmap.WallPlane[elevTileY][elevTileX-2] == 0x4c {
				doorEntity.AdditionalKeys["angle"] = "270"
			} else if elevTileY < 126 && rtlmap.WallPlane[elevTileY+2][elevTileX] == 0x4c {
				doorEntity.AdditionalKeys["angle"] = "0"
			} else if elevTileY > 1 && rtlmap.WallPlane[elevTileY-2][elevTileX] == 0x4c {
				doorEntity.AdditionalKeys["angle"] = "180"
			} else {
				log.Printf("Unknown angle for elevator tile at (%d,%d)", elevTileX, elevTileY)
				doorEntity.AdditionalKeys["angle"] = "-1"
			}
		} else {
			doorEntity.AdditionalKeys["angle"] = "-1"
		}
		doorEntity.AdditionalKeys["speed"] = "290"

		if timeBeforeOpen > 0 {
			// timed door, only open after a delayed trigger
			entityName := fmt.Sprintf("door_%d_%d", door.Tiles[0].X, door.Tiles[0].Y)
			doorEntity.AdditionalKeys["targetname"] = entityName
			doorEntity.AdditionalKeys["wait"] = "-1"
			triggerEntity := quakemap.NewEntity(0, "trigger_relay", qm)
			triggerEntity.AdditionalKeys["target"] = entityName
			triggerEntity.AdditionalKeys["targetname"] = "timed_delay_trigger"
			triggerEntity.AdditionalKeys["delay"] = fmt.Sprintf("%d", timeBeforeOpen)
			triggerEntity.AdditionalKeys["message"] = "Time-delay door opens."
			triggerEntity.OriginX = float64(door.Tiles[0].X)*gridSizeX + (gridSizeX / 2)
			triggerEntity.OriginY = float64(door.Tiles[0].Y)*-gridSizeY - (gridSizeY / 2.0)
			triggerEntity.OriginZ = floorDepth + (gridSizeZ / 2)
			qm.Entities = append(qm.Entities, triggerEntity)

			if timedTriggerEntity == nil {
				timedTriggerEntity := quakemap.NewEntity(0, "trigger_once", qm)
				timedTriggerEntity.AdditionalKeys["target"] = "timed_delay_trigger"
				timedTriggerEntity.Brushes = append(timedTriggerEntity.Brushes, quakemap.BasicCuboid(
					float64(rtlmap.SpawnX)*gridSizeX, float64(rtlmap.SpawnY)*-gridSizeY, floorDepth,
					float64(rtlmap.SpawnX+1)*gridSizeX, float64(rtlmap.SpawnY+1)*-gridSizeY, floorDepth+gridSizeZ,
					"__TB_empty", scale, false,
				))
				qm.Entities = append(qm.Entities, timedTriggerEntity)
			}
		}

		qm.Entities = append(qm.Entities, doorEntity)
	}
}

func ConvertRTLMapToQuakeMapFile(rtlmap *RTLMapData, textureWad string, scale float64, dusk bool) *quakemap.QuakeMap {

	// worldspawn:
	// 1. build 128x128 floor
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorLength float64 = gridSizeX * 128.0
	var floorWidth float64 = gridSizeY * 128.0
	var floorDepth float64 = 64.0 * scale

	var playerStartX float64 = float64(rtlmap.SpawnX)*gridSizeX + (gridSizeX / 2.0)
	var playerStartY float64 = float64(rtlmap.SpawnY)*-gridSizeY - (gridSizeY / 2.0)
	var playerAngle float64
	switch rtlmap.SpawnDirection {
	case 0: // up
		playerAngle = 90
	case 1: // right
		playerAngle = 0
	case 2: // down
		playerAngle = 270
	case 3: // left
		playerAngle = 180
	}

	qm := quakemap.NewQuakeMap(playerStartX, playerStartY, floorDepth+32)
	qm.InfoPlayerStart.Angle = playerAngle
	qm.Wad = textureWad

	floorBrush := quakemap.BasicCuboid(
		0, 0, 0,
		floorWidth, -floorLength, floorDepth,
		rtlmap.FloorTexture(),
		scale, false)
	qm.WorldSpawn.Brushes = []quakemap.Brush{floorBrush}

	// add ceiling if declared
	ceilTexture := rtlmap.CeilingTexture()
	if ceilTexture != "" {
		ceilz1 := floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
		ceilz2 := floorDepth + float64(rtlmap.FloorHeight()+1)*gridSizeZ
		ceilBrush := quakemap.BasicCuboid(
			0, 0, ceilz1,
			floorWidth, -floorLength, ceilz2,
			ceilTexture,
			scale, false)
		qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, ceilBrush)
	}

	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			wallInfo := rtlmap.ActorGrid[y][x]
			itemInfo := rtlmap.ActorGrid[y][x].Item

			switch wallInfo.Type {
			case WALL_Regular, WALL_Elevator:
				CreateRegularWall(rtlmap, x, y, scale, qm)
			case WALL_ThinWall:
				CreateThinWall(rtlmap, x, y, scale, qm)
			case WALL_AnimatedWall:
				CreateRegularWall(rtlmap, x, y, scale, qm)
			case WALL_Platform:
				CreatePlatform(rtlmap, x, y, scale, qm)
			case WALL_MaskedWall:
				CreateMaskedWall(rtlmap, x, y, scale, qm)
			}

			if itemInfo != nil {
				if itemInfo.AddCallback != nil {
					itemInfo.AddCallback(x, y, gridSizeX, gridSizeY, gridSizeZ, itemInfo, rtlmap, qm, dusk)
				} else {
					entityName := itemInfo.QuakeEntityName
					if dusk {
						entityName = itemInfo.DuskEntityName
					}
					entity := quakemap.NewEntity(0, entityName, qm)
					entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2.0)
					entity.OriginY = float64(y)*-gridSizeY - (gridSizeY / 2.0)
					if wallInfo.HeightOffset != 0 {
						entity.OriginZ = floorDepth + (gridSizeZ / 2) + ((float64(rtlmap.FloorHeight()-1) * gridSizeZ) + float64(wallInfo.HeightOffset)*(scale/64.0))
					} else {
						entity.OriginZ = floorDepth + (gridSizeZ / 2)
					}
					qm.Entities = append(qm.Entities, entity)
				}
			}

			if len(wallInfo.MapTriggers) > 0 {
				log.Printf("Creating triggers at (%d,%d)", wallInfo.X, wallInfo.Y)
				CreateTrigger(rtlmap, &wallInfo, scale, qm)
			}

		}
	}

	CreateDoorEntities(rtlmap, scale, dusk, qm)
	LinkElevators(rtlmap, textureWad, floorDepth, gridSizeX, gridSizeY, gridSizeZ, scale, dusk, qm)

	// 2. TODO: clip brushes around floor extending height
	return qm
}
