package rtl

import (
	"fmt"
	"gitlab.com/camtap/rott2quake/pkg/quakemap"
	"log"
	"math"
)

// RTL to Quake MAP conversion functions
//
// Keep in mind that in ROTT, Y increases southward while in
// Trenchbroom, Y increases northward, so Y axis values are
// inverted.

const (
	PushWallTriggerMargin float64 = 0.15
	MovingObjectBaseSpeed         = 55.0
	ElevatingGADBaseSpeed         = 150.0
)

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

func SpawnClipEntity(x1, y1, z1, x2, y2, z2 float64, actor *ActorInfo, dusk bool, qm *quakemap.QuakeMap) *quakemap.Entity {
	clipBrush := quakemap.BasicCuboid(
		x1, y1, z1,
		x2, y2, z2,
		"clip", 1.0, false)
	clipEntity := qm.SpawnEntity("func_detail", 0)
	if dusk {
		// FIXME when clip textures are no longer busted in Dusk,
		// see https://discord.com/channels/240195284695646209/586508960128040961/799031954728419348
		clipEntity.ClassName = "func_wall"
		clipEntity.AdditionalKeys["rendermode"] = "1"
		clipEntity.AdditionalKeys["renderamt"] = "0"
		clipBrush = quakemap.BasicCuboid(
			x1, y1, z1,
			x2, y2, z2,
			// just pick an arbitrary texture besides "clip"
			"FLRCL1", 1.0, false)
	}
	AddDefaultEntityKeys(clipEntity, actor)
	clipEntity.Brushes = []quakemap.Brush{clipBrush}
	return clipEntity
}

func AddDefaultEntityKeys(entity *quakemap.Entity, actor *ActorInfo) {
	entity.AdditionalKeys["_r2q_x"] = fmt.Sprintf("%d", actor.X)
	entity.AdditionalKeys["_r2q_y"] = fmt.Sprintf("%d", actor.Y)
	entity.AdditionalKeys["_r2q_wallval"] = fmt.Sprintf("%d (%04x)", actor.WallValue, actor.WallValue)
	entity.AdditionalKeys["_r2q_spriteval"] = fmt.Sprintf("%d (%04x)", actor.SpriteValue, actor.SpriteValue)
	entity.AdditionalKeys["_r2q_infoval"] = fmt.Sprintf("%d (%04x)", actor.InfoValue, actor.InfoValue)
	entity.AdditionalKeys["_r2q_tile"] = fmt.Sprintf("%d", actor.Tile)
	entity.AdditionalKeys["_r2q_type"] = actor.Type.String()
	if actor.Type == WALL_MaskedWall {
		maskedWallInfo := MaskedWalls[actor.Tile]
		entity.AdditionalKeys["_r2q_mw_flags"] = maskedWallInfo.Flags.String()
	}
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
		floor1Entity := qm.SpawnEntity("func_button", 0)
		floor1Entity.AdditionalKeys["target"] = fmt.Sprintf("elev_%d_1_trigger", linkCode)
		floor1Entity.AdditionalKeys["angle"] = button1Angle
		floor1Entity.AdditionalKeys["lip"] = "1"
		floor1Entity.AddBrush(quakemap.BasicCuboid(floor1ButtonX1, floor1ButtonY1, floorDepth,
			floor1ButtonX2, floor1ButtonY2, float64(rtlmap.FloorHeight()+1)*gridSizeZ,
			"ELEV5", scale, false))
		AddDefaultEntityKeys(floor1Entity, &elev1.Switch)

		floor1TriggerEntity := qm.SpawnEntity("trigger_teleport", 0)
		floor1TriggerEntity.AdditionalKeys["targetname"] = fmt.Sprintf("elev_%d_1_trigger", linkCode)
		floor1TriggerEntity.AdditionalKeys["target"] = fmt.Sprintf("elev_%d_1", linkCode)
		floor1TriggerEntity.AddBrush(quakemap.BasicCuboid(float64(elev1.Floor.X)*gridSizeX, float64(elev1.Floor.Y)*-gridSizeY, floorDepth,
			float64(elev1.Floor.X+1)*gridSizeX, float64(elev1.Floor.Y+1)*-gridSizeY, floorDepth+gridSizeZ,
			"__TB_empty", -1.0, false))

		floor1DestEntity := qm.SpawnEntity("info_teleport_destination", 0)
		floor1DestEntity.OriginX = float64(elev2.Floor.X)*gridSizeX + (gridSizeX / 2)
		floor1DestEntity.OriginY = float64(elev2.Floor.Y)*-gridSizeY - (gridSizeY / 2)
		floor1DestEntity.OriginZ = floorDepth
		floor1DestEntity.AdditionalKeys["targetname"] = fmt.Sprintf("elev_%d_1", linkCode)
		floor1DestEntity.AdditionalKeys["angle"] = button2Angle

		// elevator 2
		floor2Entity := qm.SpawnEntity("func_button", 0)
		floor2Entity.AdditionalKeys["target"] = fmt.Sprintf("elev_%d_2_trigger", linkCode)
		floor2Entity.AdditionalKeys["angle"] = button2Angle
		floor2Entity.AdditionalKeys["lip"] = "1"
		floor2Entity.AddBrush(quakemap.BasicCuboid(floor2ButtonX1, floor2ButtonY1, floorDepth,
			floor2ButtonX2, floor2ButtonY2, float64(rtlmap.FloorHeight()+1)*gridSizeZ,
			"ELEV5", scale, false))
		AddDefaultEntityKeys(floor2Entity, &elev2.Switch)

		floor2TriggerEntity := qm.SpawnEntity("trigger_teleport", 0)
		floor2TriggerEntity.AdditionalKeys["targetname"] = fmt.Sprintf("elev_%d_2_trigger", linkCode)
		floor2TriggerEntity.AdditionalKeys["target"] = fmt.Sprintf("elev_%d_2", linkCode)
		floor2TriggerEntity.AddBrush(quakemap.BasicCuboid(float64(elev2.Floor.X)*gridSizeX, float64(elev2.Floor.Y)*-gridSizeY, floorDepth,
			float64(elev2.Floor.X+1)*gridSizeX, float64(elev2.Floor.Y+1)*-gridSizeY, floorDepth+gridSizeZ,
			"__TB_empty", -1.0, false))

		floor2DestEntity := qm.SpawnEntity("info_teleport_destination", 0)
		floor2DestEntity.OriginX = float64(elev1.Floor.X)*gridSizeX + (gridSizeX / 2)
		floor2DestEntity.OriginY = float64(elev1.Floor.Y)*-gridSizeY - (gridSizeY / 2)
		floor2DestEntity.OriginZ = floorDepth
		floor2DestEntity.AdditionalKeys["targetname"] = fmt.Sprintf("elev_%d_2", linkCode)
		floor2DestEntity.AdditionalKeys["angle"] = button1Angle
	}
}

func CreateGAD(rtlmap *RTLMapData, actor *ActorInfo, scale float64, dusk bool, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	dX := float64(actor.X)*gridSizeX + (gridSizeX / 2.0)
	dY := float64(actor.Y)*-gridSizeY - (gridSizeY / 2.0)
	zOffset := rtlmap.ZOffset(actor.InfoValue, scale)
	dZ := floorDepth + zOffset

	var gadBrushes []quakemap.Brush
	for _, brush := range quakemap.GADBrushes {
		newBrush := brush.Clone()
		newBrush.Translate(dX, dY, dZ)
		gadBrushes = append(gadBrushes, newBrush)
	}
	entityClassname := "func_detail"
	entityKeys := make(map[string]string)

	clipBrush := gadBrushes[len(gadBrushes)-1]
	if dusk {
		// HACK/FIXME: Dusk SDK ignores clip textures, so make them separate
		// entities. Assume last brush is the surrounding clip texture.
		clipBrush.SetTextureForAllPlanes("FLRCL1")
		clipEntity := qm.SpawnEntity("func_wall", 0)
		clipEntity.AdditionalKeys["rendermode"] = "1"
		clipEntity.AdditionalKeys["renderamt"] = "0"
		clipEntity.AddBrush(clipBrush)
		gadBrushes = gadBrushes[0 : len(gadBrushes)-1]
	}

	switch actor.SpriteValue {
	case MovingGADEast, MovingGADNorth, MovingGADWest, MovingGADSouth:
		// build trackpath
		entityClassname = "func_train"
		moveInfo := MoveWallSpriteIDs[actor.SpriteValue]

		var lastPathCorner, currentPathCorner *quakemap.Entity
		pathType, gadPath, numNodes := rtlmap.DetermineWallPath(actor, false)

		if pathType != PATH_Perpetual {
			log.Panicf("GAD at (%d,%d) not perpetual?", actor.X, actor.Y)
		}
		initialCorner := quakemap.NewEntity(0, "path_corner", qm)
		initialCorner.OriginX = dX - clipBrush.Width()/2.0
		initialCorner.OriginY = dY - clipBrush.Length()/2.0
		initialCorner.OriginZ = dZ
		initialCorner.AdditionalKeys["targetname"] = fmt.Sprintf("gadpath_%d_%d_init", actor.X, actor.Y)
		initialCorner.AdditionalKeys["wait"] = "0.00001"
		lastPathCorner = initialCorner
		qm.AddEntity(lastPathCorner)
		entityKeys["target"] = initialCorner.AdditionalKeys["targetname"]
		entityKeys["speed"] = fmt.Sprintf("%.02f", float64(moveInfo.Speed)*MovingObjectBaseSpeed*scale)

		currentNode := gadPath
		nodeToTargetNames := make(map[*PathNode]string)
		for i := 0; i < numNodes; i++ {
			currentPathCorner = qm.SpawnEntity("path_corner", 0)
			currentPathCorner.OriginZ = dZ
			currentPathCorner.OriginX = (float64(currentNode.X))*gridSizeX + (gridSizeX / 2.0) - clipBrush.Width()/2.0
			currentPathCorner.OriginY = (float64(currentNode.Y))*-gridSizeY - (gridSizeY / 2.0) - clipBrush.Length()/2.0
			targetName := fmt.Sprintf("gadpath_%d_%d_%d", actor.X, actor.Y, i)
			nodeToTargetNames[currentNode] = targetName
			currentPathCorner.AdditionalKeys["targetname"] = targetName
			currentPathCorner.AdditionalKeys["wait"] = "0.00001"
			lastPathCorner.AdditionalKeys["target"] = targetName
			currentNode = currentNode.Next
			lastPathCorner = currentPathCorner
		}
		currentPathCorner.AdditionalKeys["target"] = nodeToTargetNames[currentNode]
	case ElevatingGAD:
		// build single-column trackpath
		entityClassname = "func_train"
		upperPathEntityName := fmt.Sprintf("gad_%d_%d_upper", actor.X, actor.Y)
		lowerPathEntityName := fmt.Sprintf("gad_%d_%d_lower", actor.X, actor.Y)

		upperPathEntity := quakemap.NewEntity(0, "path_corner", qm)
		upperPathEntity.OriginX = dX - clipBrush.Width()/2.0
		upperPathEntity.OriginY = dY - clipBrush.Length()/2.0
		upperPathEntity.OriginZ = dZ
		upperPathEntity.AdditionalKeys["target"] = lowerPathEntityName
		upperPathEntity.AdditionalKeys["targetname"] = upperPathEntityName
		upperPathEntity.AdditionalKeys["wait"] = "1"

		lowerPathEntity := quakemap.NewEntity(0, "path_corner", qm)
		lowerPathEntity.OriginX = dX - clipBrush.Width()/2.0
		lowerPathEntity.OriginY = dY - clipBrush.Length()/2.0
		lowerPathEntity.OriginZ = floorDepth
		lowerPathEntity.AdditionalKeys["target"] = upperPathEntityName
		lowerPathEntity.AdditionalKeys["targetname"] = lowerPathEntityName
		lowerPathEntity.AdditionalKeys["wait"] = "1"
		entityKeys["target"] = upperPathEntityName
		entityKeys["speed"] = fmt.Sprintf("%.02f", ElevatingGADBaseSpeed*scale)

		qm.AddEntity(upperPathEntity)
		qm.AddEntity(lowerPathEntity)
	}

	GADEntity := qm.SpawnEntity(entityClassname, 0)
	GADEntity.Brushes = gadBrushes
	AddDefaultEntityKeys(GADEntity, actor)
	GADEntity.AdditionalKeys["_r2q_zoffset"] = fmt.Sprintf("%.02f", rtlmap.ZOffset(actor.InfoValue, scale))
	for k, v := range entityKeys {
		GADEntity.AdditionalKeys[k] = v
	}

	if actor.SpriteValue == StaticGAD {
		// add clip textures around other static GADs that are less
		// than 7 units taller/shorter to allow stepping onto them
		// without jumping

		Xmargin := (gridSizeX - GADEntity.Width()) / 2.0
		Ymargin := (gridSizeY - GADEntity.Length()) / 2.0

		stepZCoords := func(neighborZOffset, zDiff float64) (float64, float64) {
			var stepZ1, stepZ2 float64
			switch {
			case zDiff > 0:
				stepZ1 = dZ + (GADEntity.Height() / 2.0)
				stepZ2 = stepZ1 + (zDiff / 2.0)
			case zDiff == 0:
				stepZ1 = dZ + GADEntity.Height() - 1
				stepZ2 = stepZ1 + 1
			case zDiff < 0:
				stepZ1 = floorDepth + neighborZOffset + (GADEntity.Height() / 2.0) - (zDiff / 2.0)
				stepZ2 = stepZ1 - (zDiff / 2.0)
			}
			return stepZ1, stepZ2
		}

		maxStepHeight := 24 * scale

		// east
		if actor.X < 127 && rtlmap.ActorGrid[actor.Y][actor.X+1].SpriteValue == StaticGAD {
			neighbor := rtlmap.ActorGrid[actor.Y][actor.X+1]
			neighborZOffset := rtlmap.ZOffset(neighbor.InfoValue, scale)
			zDiff := neighborZOffset - zOffset
			if math.Abs(zDiff) <= maxStepHeight {
				log.Printf("(%d,%d)->(%d,%d) adding clip to east", actor.X, actor.Y, neighbor.X, neighbor.Y)
				stepX1 := (float64(actor.X) * gridSizeX) + GADEntity.Width() + Xmargin
				stepX2 := stepX1 + Xmargin
				stepY1 := (float64(actor.Y) * -gridSizeY) - Ymargin
				stepY2 := stepY1 - GADEntity.Length()
				stepZ1, stepZ2 := stepZCoords(neighborZOffset, zDiff)
				log.Printf("(%d,%d)->(%d,%d) stepZ1 = %.02f, stepZ2 = %.02f", actor.X, actor.Y, neighbor.X, neighbor.Y,
					stepZ1, stepZ2)
				_ = SpawnClipEntity(stepX1, stepY1, stepZ1, stepX2, stepY2, stepZ2,
					actor, dusk, qm)
			}
		}

		// north
		if actor.Y > 0 && rtlmap.ActorGrid[actor.Y-1][actor.X].SpriteValue == StaticGAD {
			neighbor := rtlmap.ActorGrid[actor.Y-1][actor.X]
			neighborZOffset := rtlmap.ZOffset(neighbor.InfoValue, scale)
			zDiff := neighborZOffset - zOffset
			if math.Abs(zDiff) <= maxStepHeight {
				stepY1 := float64(actor.Y) * -gridSizeY
				stepY2 := stepY1 - Ymargin
				stepX1 := (float64(actor.X) * gridSizeX) + Xmargin
				stepX2 := stepX1 + GADEntity.Width()
				stepZ1, stepZ2 := stepZCoords(neighborZOffset, zDiff)
				_ = SpawnClipEntity(stepX1, stepY1, stepZ1, stepX2, stepY2, stepZ2,
					actor, dusk, qm)
			}
		}

		// west
		if actor.X > 0 && rtlmap.ActorGrid[actor.Y][actor.X-1].SpriteValue == StaticGAD {
			neighbor := rtlmap.ActorGrid[actor.Y][actor.X-1]
			neighborZOffset := rtlmap.ZOffset(neighbor.InfoValue, scale)
			zDiff := neighborZOffset - zOffset
			if math.Abs(zDiff) <= maxStepHeight {
				stepX1 := float64(actor.X) * gridSizeX
				stepX2 := stepX1 + Xmargin
				stepY1 := (float64(actor.Y) * -gridSizeY) - Ymargin
				stepY2 := stepY1 - GADEntity.Length()
				stepZ1, stepZ2 := stepZCoords(neighborZOffset, zDiff)
				_ = SpawnClipEntity(stepX1, stepY1, stepZ1, stepX2, stepY2, stepZ2,
					actor, dusk, qm)
			}
		}

		// south
		if actor.Y < 127 && rtlmap.ActorGrid[actor.Y+1][actor.X].SpriteValue == StaticGAD {
			neighbor := rtlmap.ActorGrid[actor.Y+1][actor.X]
			neighborZOffset := rtlmap.ZOffset(neighbor.InfoValue, scale)
			zDiff := neighborZOffset - zOffset
			if math.Abs(zDiff) <= maxStepHeight {
				stepY1 := float64(actor.Y)*-gridSizeY - Ymargin - GADEntity.Length()
				stepY2 := stepY1 - Ymargin
				stepX1 := (float64(actor.X) * gridSizeX) + Xmargin
				stepX2 := stepX1 + GADEntity.Width()
				stepZ1, stepZ2 := stepZCoords(neighborZOffset, zDiff)
				_ = SpawnClipEntity(stepX1, stepY1, stepZ1, stepX2, stepY2, stepZ2,
					actor, dusk, qm)
			}
		}

	}
}

func ClipHeight(rtlmap *RTLMapData, actor *ActorInfo, scale float64) float64 {
	switch actor.Type {
	case WALL_Platform:
		switch actor.InfoValue {
		case 1, 8, 9:
			return scale*64.0 + float64(rtlmap.FloorHeight()-1)*(scale*64.0)
		case 5, 6:
			return (scale * 64.0) * 2.0
		default:
			return 0.0
		}
	case SPR_GAD:
		return (scale * 64.0) + rtlmap.ZOffset(actor.InfoValue, scale)
	default:
		return 0.0
	}
}

func AddThinWallClipTextures(rtlmap *RTLMapData, actor *ActorInfo, scale float64, dusk bool, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale

	wallDirection, _, _ := rtlmap.ThinWallDirection(actor.X, actor.Y)

	// add clip textures to prevent the player from falling in
	// between the face of a thin wall and another object
	if wallDirection == WALLDIR_NorthSouth {
		westClipZ := ClipHeight(rtlmap, &rtlmap.ActorGrid[actor.Y][actor.X-1], scale)
		if westClipZ > 0.0 {
			// clip tile to west
			_ = SpawnClipEntity(
				float64(actor.X)*gridSizeX,
				float64(actor.Y)*-gridSizeY,
				westClipZ,
				(float64(actor.X)+0.5)*gridSizeX,
				float64(actor.Y+1)*-gridSizeY,
				westClipZ-1,
				actor, dusk, qm,
			)
		}

		eastClipZ := ClipHeight(rtlmap, &rtlmap.ActorGrid[actor.Y][actor.X+1], scale)
		if eastClipZ > 0.0 {
			// clip tile to east
			_ = SpawnClipEntity(
				(float64(actor.X)+0.5)*gridSizeX,
				float64(actor.Y)*-gridSizeY,
				eastClipZ,
				float64(actor.X+1)*gridSizeX,
				float64(actor.Y+1)*-gridSizeY,
				eastClipZ-1,
				actor, dusk, qm,
			)
		}
	} else {
		northClipZ := ClipHeight(rtlmap, &rtlmap.ActorGrid[actor.Y-1][actor.X], scale)
		if northClipZ > 0.0 {
			// clip tile to north
			_ = SpawnClipEntity(
				float64(actor.X)*gridSizeX,
				float64(actor.Y)*-gridSizeY,
				northClipZ,
				float64(actor.X+1)*gridSizeX,
				(float64(actor.Y)+0.5)*-gridSizeY,
				northClipZ-1,
				actor, dusk, qm,
			)
		}

		southClipZ := ClipHeight(rtlmap, &rtlmap.ActorGrid[actor.Y+1][actor.X], scale)
		if southClipZ > 0.0 {
			// clip tile to south
			_ = SpawnClipEntity(
				float64(actor.X)*gridSizeX,
				(float64(actor.Y)+0.5)*-gridSizeY,
				southClipZ,
				float64(actor.X+1)*gridSizeX,
				float64(actor.Y+1)*-gridSizeY,
				southClipZ-1,
				actor, dusk, qm,
			)
		}
	}
}

func CreateThinWall(rtlmap *RTLMapData, x, y int, scale float64, dusk bool, qm *quakemap.QuakeMap) {
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
			qm.WorldSpawn.AddBrush(wallColumn)
		case 4:
			// above only
			var z1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			var z2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.AddBrush(wallColumn)
		case 5:
			// below only
			var z1 float64 = floorDepth
			var z2 float64 = floorDepth + gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.AddBrush(wallColumn)
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
			qm.WorldSpawn.AddBrush(wallColumn1)
			qm.WorldSpawn.AddBrush(wallColumn2)
		case 7:
			// everything but below
			var z1 float64 = floorDepth + gridSizeZ
			var z2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.AddBrush(wallColumn)
		case 8:
			// middle only
			var z1 float64 = floorDepth + gridSizeZ
			var z2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.AddBrush(wallColumn)
		case 9:
			// everything but above
			var z1 float64 = floorDepth
			var z2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
				texName, scale, false)
			qm.WorldSpawn.AddBrush(wallColumn)
		}

		AddThinWallClipTextures(rtlmap, &actor, scale, dusk, qm)
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
	triggerEntity := qm.SpawnEntity("trigger_multiple", 0)
	triggerEntity.AdditionalKeys["target"] = fmt.Sprintf("trigger_%d_%d", actor.X, actor.Y)
	triggerEntity.AdditionalKeys["message"] = "Switch Triggered."
	triggerEntity.AddBrush(wallColumnBrush)
	AddDefaultEntityKeys(triggerEntity, actor)
}

func CreateTouchplate(rtlmap *RTLMapData, actor *ActorInfo, scale float64, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	relayTargetName := fmt.Sprintf("trigger_%d_%d", actor.X, actor.Y)

	triggerEntity := qm.SpawnEntity("trigger_once", 0)
	triggerEntity.AdditionalKeys["target"] = relayTargetName
	triggerEntity.AdditionalKeys["message"] = "Touchplate Triggered"
	triggerEntity.AddBrush(
		quakemap.BasicCuboid(float64(actor.X)*gridSizeX, float64(actor.Y)*-gridSizeY, floorDepth,
			float64(actor.X+1)*gridSizeX, float64(actor.Y+1)*-gridSizeY, floorDepth+gridSizeZ,
			"__TB_empty", scale, false))
	AddDefaultEntityKeys(triggerEntity, actor)
}

func CreateSingleUnitWall(rtlmap *RTLMapData, x, y int, scale float64, textureName string, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	x1 := float64(x) * gridSizeX
	y1 := float64(y) * -gridSizeY
	x2 := float64(x+1) * gridSizeX
	y2 := float64(y+1) * -gridSizeY

	gatez1 := floorDepth
	gatez2 := floorDepth + gridSizeZ
	wallz1 := gatez2 + 1
	wallz2 := floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ

	gateColumn := quakemap.BasicCuboid(x1, y1, gatez1,
		x2, y2, gatez2,
		textureName, scale, false)
	wallColumn := quakemap.BasicCuboid(x1, y1, wallz1,
		x2, y2, wallz2,
		"WALL22", scale, false)
	qm.WorldSpawn.AddBrush(gateColumn)
	qm.WorldSpawn.AddBrush(wallColumn)
}

func CreateRegularWall(rtlmap *RTLMapData, x, y int, scale float64, qm *quakemap.QuakeMap) {
	switch rtlmap.WallPlane[y][x] {
	case 0x2f:
		CreateSingleUnitWall(rtlmap, x, y, scale, "EXIT", qm)
	case 0x30:
		CreateSingleUnitWall(rtlmap, x, y, scale, "ENTRANCE", qm)
	default:
		CreateRegularWallSingleTexture(rtlmap, x, y, scale, qm)
	}
}

func CreateRegularWallSingleTexture(rtlmap *RTLMapData, x, y int, scale float64, qm *quakemap.QuakeMap) {
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
	var wallColumn quakemap.Brush

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
	wallColumn = quakemap.BasicCuboid(x1, y1, z1,
		x2, y2, z2,
		texName, scale, true)

	if actor.MapFlags&WALLFLAGS_Moving != 0 {
		var lastPathCorner, currentPathCorner *quakemap.Entity
		pathType, wallPath, numNodes := rtlmap.DetermineWallPath(&actor, (spriteVal < 256))
		moveWallInfo = MoveWallSpriteIDs[spriteVal]
		initialCorner = quakemap.NewEntity(0, "path_corner", qm)
		cornerZ := floorDepth
		initialCorner.OriginX = (float64(x)) * gridSizeX
		initialCorner.OriginY = (float64(y) + 1) * -gridSizeY
		initialCorner.OriginZ = cornerZ
		initialCorner.AdditionalKeys["targetname"] = fmt.Sprintf("movewallpath_%d_%d_init", actor.X, actor.Y)
		initialCorner.AdditionalKeys["wait"] = "0.00001"
		lastPathCorner = initialCorner
		qm.AddEntity(lastPathCorner)

		currentNode := wallPath
		nodeToTargetNames := make(map[*PathNode]string)
		for i := 0; i < numNodes; i++ {
			currentPathCorner = qm.SpawnEntity("path_corner", 0)
			currentPathCorner.OriginZ = cornerZ
			currentPathCorner.OriginX = (float64(currentNode.X)) * gridSizeX
			currentPathCorner.OriginY = (float64(currentNode.Y) + 1) * -gridSizeY
			targetName := fmt.Sprintf("movewallpath_%d_%d_%d", actor.X, actor.Y, i)
			nodeToTargetNames[currentNode] = targetName
			currentPathCorner.AdditionalKeys["targetname"] = targetName
			currentPathCorner.AdditionalKeys["wait"] = "0.00001"
			lastPathCorner.AdditionalKeys["target"] = targetName
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
		qm.WorldSpawn.AddBrush(wallColumn)
	} else {
		entity := qm.SpawnEntity(entityType, 0)
		entity.Brushes = []quakemap.Brush{wallColumn}
		entity.AdditionalKeys["_x"] = fmt.Sprintf("%d", actor.X)
		entity.AdditionalKeys["_y"] = fmt.Sprintf("%d", actor.Y)
		AddDefaultEntityKeys(entity, &actor)
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
				relayEntity := qm.SpawnEntity("trigger_relay", 0)
				relayEntity.OriginX = (float64(actor.X) + 0.5) * gridSizeX
				relayEntity.OriginY = (float64(actor.Y) + 0.5) * -gridSizeY
				relayEntity.OriginZ = floorDepth + (float64(rtlmap.FloorHeight()+1))*gridSizeZ
				relayEntity.AdditionalKeys["targetname"] = fmt.Sprintf("trigger_%d_%d", triggerX, triggerY)
				relayEntity.AdditionalKeys["target"] = wallTargetName
			} else if spriteVal < 256 {
				var tx1, ty1, tx2, ty2 float64

				// only allow pushing from the opposite direction it
				// moves toward when triggered
				switch actor.SpriteValue {
				case 300, uint16(DIR_East) + uint16(ICONARROWS):
					tx1 = (float64(actor.X) * gridSizeX) - 1
					tx2 = tx1 + 1.0
					ty1 = (float64(actor.Y) + PushWallTriggerMargin) * -gridSizeY
					ty2 = (float64(actor.Y+1) - PushWallTriggerMargin) * -gridSizeY
				case 318, uint16(DIR_North) + uint16(ICONARROWS):
					tx1 = (float64(actor.X) + PushWallTriggerMargin) * gridSizeX
					tx2 = (float64(actor.X+1) - PushWallTriggerMargin) * gridSizeX
					ty1 = (float64(actor.Y+1) * -gridSizeY) + 1
					ty2 = ty1 - 1.0
				case 336, uint16(DIR_West) + uint16(ICONARROWS):
					tx1 = (float64(actor.X+1) * gridSizeX) - 1
					tx2 = tx1 + 1.0
					ty1 = (float64(actor.Y) + PushWallTriggerMargin) * -gridSizeY
					ty2 = (float64(actor.Y+1) - PushWallTriggerMargin) * -gridSizeY
				case 354, uint16(DIR_South) + uint16(ICONARROWS):
					tx1 = (float64(actor.X) + PushWallTriggerMargin) * gridSizeX
					tx2 = (float64(actor.X+1) - PushWallTriggerMargin) * gridSizeX
					ty1 = (float64(actor.Y) * -gridSizeY) + 1
					ty2 = ty1 - 1.0
				default:
					panic("yes you're stuck implementing diagonal pushwall triggers")
				}
				// add pushwall trigger_once entity within the wall
				pushWallTriggerEntity := qm.SpawnEntity("trigger_once", 0)
				pushWallTriggerEntity.AddBrush(
					quakemap.BasicCuboid(tx1, ty1, z1, tx2, ty2, z2, "trigger", scale, true),
				)
				pushWallTriggerEntity.AdditionalKeys["_x"] = fmt.Sprintf("%d", actor.X)
				pushWallTriggerEntity.AdditionalKeys["_y"] = fmt.Sprintf("%d", actor.Y)
				pushWallTriggerEntity.AdditionalKeys["target"] = wallTargetName
				pushWallTriggerEntity.AdditionalKeys["targetname"] = fmt.Sprintf("movewallpath_%d_%d_push", actor.X, actor.Y)
				entity.AdditionalKeys["targetname"] = fmt.Sprintf("movewallpath_%d_%d_wall", actor.X, actor.Y)
			}
			entity.AdditionalKeys["speed"] = fmt.Sprintf("%.02f", float64(moveWallInfo.Speed)*MovingObjectBaseSpeed*scale)
		}

		if entityType == "func_wall" && actor.Damage {
			// add trigger_hurt over wall to mimic a damaging wall
			hurtEntity := qm.SpawnEntity("trigger_hurt", 0)
			hurtEntity.Brushes = append(hurtEntity.Brushes, quakemap.BasicCuboid(x1, y1, z1,
				x2, y2, z2,
				texName, scale, true))
			hurtEntity.AdditionalKeys["_x"] = fmt.Sprintf("%d", actor.X)
			hurtEntity.AdditionalKeys["_y"] = fmt.Sprintf("%d", actor.Y)
			AddDefaultEntityKeys(hurtEntity, &actor)
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
			aboveEntity := qm.SpawnEntity(aboveClassName, 0)
			aboveEntity.AddBrush(aboveColumn)
			AddDefaultEntityKeys(aboveEntity, &actor)
		}

		// middle
		if platformInfo.Middle != "" && floorHeight > 2 {
			var middlez1 float64 = floorDepth + gridSizeZ
			var middlez2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			mwColumn := quakemap.BasicCuboid(x1, y1, middlez1, x2, y2, middlez2,
				"{"+platformInfo.Middle,
				scale, false)
			qm.WorldSpawn.AddBrush(mwColumn)
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
			bottomEntity := qm.SpawnEntity(className, 0)
			bottomEntity.AddBrush(column)
			AddDefaultEntityKeys(bottomEntity, &actor)
		}
	}
}

func CreateMaskedWall(rtlmap *RTLMapData, x, y int, scale float64, dusk bool, qm *quakemap.QuakeMap) {
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

		xScaleFactor := 1.0

		// see if it's part of a MW_Multi masked wall setup and
		// needs to be flipped around
		checkAdjacentMultiWall := func(ax, ay int, checkGreater bool) bool {
			adjacentActor := rtlmap.ActorGrid[ay][ax]
			if adjacentActor.Type == WALL_MaskedWall {
				adjacentWallInfo := MaskedWalls[adjacentActor.Tile]
				if adjacentWallInfo.Flags&MWF_Multi != 0 {
					if checkGreater {
						if adjacentActor.Tile > wallInfo.Tile {
							xScaleFactor = -1.0
						}
					} else {
						if adjacentActor.Tile < wallInfo.Tile {
							xScaleFactor = -1.0
						}
					}
					return true
				}
			}
			return false
		}

		if wallDirection == WALLDIR_NorthSouth {
			x1 = float64(x)*gridSizeX + (gridSizeX / 2)
			x2 = float64(x)*gridSizeX + (gridSizeX / 2) + 1
			y1 = float64(y) * -gridSizeY
			y2 = float64(y+1) * -gridSizeY

			if maskedWallInfo.Flags&MWF_Multi != 0 {
				if y > 0 {
					for ay := y - 1; ay >= 0 && checkAdjacentMultiWall(x, ay, false); ay-- {
					}
				}
				if y < 127 {
					for ay := y + 1; ay <= 127 && checkAdjacentMultiWall(x, ay, true); ay++ {
					}
				}
			}
		} else {
			x1 = float64(x) * gridSizeX
			x2 = float64(x+1) * gridSizeX
			y1 = float64(y)*-gridSizeY - (gridSizeY / 2.0)
			y2 = float64(y)*-gridSizeY - (gridSizeY / 2.0) + 1

			if maskedWallInfo.Flags&MWF_Multi != 0 {
				if x > 0 {
					for ax := x - 1; ax >= 0 && checkAdjacentMultiWall(ax, y, true); ax-- {
					}
				}
				if x < 127 {
					for ax := x + 1; ax <= 127 && checkAdjacentMultiWall(ax, y, false); ax++ {
					}
				}
			}
		}

		// above as separate entity
		if maskedWallInfo.Above != "" && floorHeight > 1 {
			var abovez1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
			var abovez2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
			aboveClassName := ClassNameForMaskedWall(&maskedWallInfo, "above")
			cuboidParams := quakemap.BasicCuboidParams("{"+maskedWallInfo.Above, scale, false)
			cuboidParams.North.TexScaleX *= xScaleFactor
			cuboidParams.South.TexScaleX *= xScaleFactor
			cuboidParams.East.TexScaleX *= xScaleFactor
			cuboidParams.West.TexScaleX *= xScaleFactor
			aboveColumn := quakemap.BuildCuboidBrush(x1, y1, abovez1, x2, y2, abovez2, cuboidParams)
			aboveEntity := qm.SpawnEntity(aboveClassName, 0)
			aboveEntity.AddBrush(aboveColumn)
			if maskedWallInfo.IsSwitch {
				aboveEntity.AdditionalKeys["target"] = fmt.Sprintf("trigger_%d_%d", wallInfo.X, wallInfo.Y)
				aboveEntity.AdditionalKeys["lip"] = fmt.Sprintf("%.02f", 64.0*scale)
			}
			AddDefaultEntityKeys(aboveEntity, &wallInfo)
		}

		// middle
		if maskedWallInfo.Middle != "" && floorHeight > 2 {
			var middlez1 float64 = floorDepth + gridSizeZ
			var middlez2 float64 = floorDepth + float64(floorHeight-1)*gridSizeZ
			middleClassName := ClassNameForMaskedWall(&maskedWallInfo, "middle")
			cuboidParams := quakemap.BasicCuboidParams("{"+maskedWallInfo.Middle, scale, false)
			cuboidParams.North.TexScaleX *= xScaleFactor
			cuboidParams.South.TexScaleX *= xScaleFactor
			cuboidParams.East.TexScaleX *= xScaleFactor
			cuboidParams.West.TexScaleX *= xScaleFactor
			mwColumn := quakemap.BuildCuboidBrush(x1, y1, middlez1, x2, y2, middlez2, cuboidParams)
			middleEntity := qm.SpawnEntity(middleClassName, 0)
			middleEntity.AddBrush(mwColumn)
			AddDefaultEntityKeys(middleEntity, &wallInfo)
		}

		// bottom
		// above as separate entity
		if maskedWallInfo.Bottom != "" {
			var z1 float64 = floorDepth
			var z2 float64 = floorDepth + gridSizeZ
			className := ClassNameForMaskedWall(&maskedWallInfo, "bottom")
			cuboidParams := quakemap.BasicCuboidParams("{"+maskedWallInfo.Bottom, scale, false)
			cuboidParams.North.TexScaleX *= xScaleFactor
			cuboidParams.South.TexScaleX *= xScaleFactor
			cuboidParams.East.TexScaleX *= xScaleFactor
			cuboidParams.West.TexScaleX *= xScaleFactor
			column := quakemap.BuildCuboidBrush(x1, y1, z1, x2, y2, z2, cuboidParams)
			bottomEntity := qm.SpawnEntity(className, 0)
			bottomEntity.AddBrush(column)
			AddDefaultEntityKeys(bottomEntity, &wallInfo)
		}

		// TODO: sides

		AddThinWallClipTextures(rtlmap, &wallInfo, scale, dusk, qm)

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
	keyCount := 0
	availKeys := []string{"item_key1", "item_key2"}
	if dusk {
		availKeys = []string{"key_red_key", "key_blue_key", "key_yellow_key"}
	}
	keyMap := make(map[DoorLock]int)

	var timedTriggerEntity *quakemap.Entity

	for doornum, door := range rtlmap.GetDoors() {
		doorEntity := qm.SpawnEntity("func_door", 0)
		timeBeforeOpen := 0
		flipTextures := false
		if door.Direction == WALLDIR_NorthSouth {
			if door.Tiles[0].Y > door.Tiles[len(door.Tiles)-1].Y &&
				door.Tiles[0].Tile > door.Tiles[len(door.Tiles)-1].Tile {
				flipTextures = true
			} else if door.Tiles[0].Y < door.Tiles[len(door.Tiles)-1].Y &&
				door.Tiles[0].Tile < door.Tiles[len(door.Tiles)-1].Tile {
				flipTextures = true
			}
		} else {
			if door.Tiles[0].X > door.Tiles[len(door.Tiles)-1].X &&
				door.Tiles[0].Tile < door.Tiles[len(door.Tiles)-1].Tile {
				flipTextures = true
			} else if door.Tiles[0].X < door.Tiles[len(door.Tiles)-1].X &&
				door.Tiles[0].Tile > door.Tiles[len(door.Tiles)-1].Tile {
				flipTextures = true
			}
		}
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
			cuboidParams := quakemap.BasicCuboidParams(texInfo.BaseTexture, scale, false)
			if flipTextures {
				cuboidParams.North.TexScaleX *= -1.0
				cuboidParams.South.TexScaleX *= -1.0
				cuboidParams.East.TexScaleX *= -1.0
				cuboidParams.West.TexScaleX *= -1.0
			}
			doorEntity.AddBrush(
				quakemap.BuildCuboidBrush(x1, y1, z1, x2, y2, z2, cuboidParams),
			)
			AddDefaultEntityKeys(doorEntity, &doorTile)
			aboveBrush := quakemap.BasicCuboid(abovex1, abovey1, z2,
				abovex2, abovey2, floorDepth+float64(rtlmap.FloorHeight())*gridSizeZ,
				texInfo.AltTexture,
				scale, false)
			qm.WorldSpawn.AddBrush(aboveBrush)
		}

		doorEntity.AdditionalKeys["_r2q_grid_start_x"] = fmt.Sprintf("%d", door.Tiles[0].X)
		doorEntity.AdditionalKeys["_r2q_grid_start_y"] = fmt.Sprintf("%d", door.Tiles[0].Y)

		if door.Lock != LOCK_Unlocked && door.Lock != LOCK_Trigger {
			if _, ok := keyMap[door.Lock]; !ok {
				// place keys on the map
				keyToUse := keyCount % len(availKeys)
				log.Printf("Using key entity %s as %s", availKeys[keyToUse], door.Lock.KeyName())
				if keyCount == len(availKeys) {
					log.Printf("More than %d keys used, this map may not be playable (or fun)", len(availKeys))
				}

				for y := 0; y < 128; y++ {
					for x := 0; x < 128; x++ {
						if rtlmap.ActorGrid[y][x].Type != WALL_Door && rtlmap.SpritePlane[y][x] == uint16(door.Lock+0x1c) {
							entity := qm.SpawnEntity(availKeys[keyToUse], 0)
							entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2)
							entity.OriginY = float64(y)*-gridSizeY - (gridSizeY / 2.0)
							entity.OriginZ = floorDepth + (gridSizeZ / 2)
							if dusk {
								// FIXME when/if dusk SDK fixes keys
								// being placed a lot lower than the
								// entity's origin
								entity.OriginZ += 600.0
							}
						}
					}
				}

				keyMap[door.Lock] = keyToUse
				keyCount += 1
			}

			if dusk {
				doorEntity.AdditionalKeys["key"] = fmt.Sprintf("%d", keyMap[door.Lock]+1)
			} else {
				doorEntity.SpawnFlags |= (2 - (keyMap[door.Lock])) * 8
			}
		}

		doorEntity.AdditionalKeys["_r2q_doornum"] = fmt.Sprintf("%d", doornum)
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
		doorEntity.AdditionalKeys["speed"] = "290.0"

		if timeBeforeOpen > 0 {
			// timed door, only open after a delayed trigger
			entityName := fmt.Sprintf("door_%d_%d", door.Tiles[0].X, door.Tiles[0].Y)
			doorEntity.AdditionalKeys["targetname"] = entityName
			doorEntity.AdditionalKeys["wait"] = "-1"
			triggerEntity := qm.SpawnEntity("trigger_relay", 0)
			triggerEntity.AdditionalKeys["target"] = entityName
			triggerEntity.AdditionalKeys["targetname"] = "timed_delay_trigger"
			triggerEntity.AdditionalKeys["delay"] = fmt.Sprintf("%d", timeBeforeOpen)
			triggerEntity.AdditionalKeys["message"] = "Time-delay door opens."
			triggerEntity.OriginX = float64(door.Tiles[0].X)*gridSizeX + (gridSizeX / 2)
			triggerEntity.OriginY = float64(door.Tiles[0].Y)*-gridSizeY - (gridSizeY / 2.0)
			triggerEntity.OriginZ = floorDepth + (gridSizeZ / 2)

			if timedTriggerEntity == nil {
				timedTriggerEntity := qm.SpawnEntity("trigger_once", 0)
				timedTriggerEntity.AdditionalKeys["target"] = "timed_delay_trigger"
				timedTriggerEntity.AddBrush(quakemap.BasicCuboid(
					float64(rtlmap.SpawnX)*gridSizeX, float64(rtlmap.SpawnY)*-gridSizeY, floorDepth,
					float64(rtlmap.SpawnX+1)*gridSizeX, float64(rtlmap.SpawnY+1)*-gridSizeY, floorDepth+gridSizeZ,
					"trigger", scale, false,
				))
			}
		}
	}
}

func AddExitPoints(rtlmap *RTLMapData, scale float64, dusk bool, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	for _, point := range rtlmap.ExitPoints {
		log.Printf("Adding map%03d exit point at (%d,%d)", point.DestMap, point.X, point.Y)
		brush := quakemap.BasicCuboid(
			(float64(point.X)+0.25)*gridSizeX,
			(float64(point.Y)+0.25)*-gridSizeY,
			floorDepth,
			(float64(point.X)+0.75)*gridSizeX,
			(float64(point.Y)+0.75)*-gridSizeY,
			floorDepth+gridSizeZ,
			"trigger", scale, false)
		entity := qm.SpawnEntity("trigger_changelevel", 0)
		entity.AdditionalKeys["map"] = fmt.Sprintf("map%03d", point.DestMap)
		entity.AddBrush(brush)
	}
}

func AddEnemies(rtlmap *RTLMapData, scale float64, dusk bool, qm *quakemap.QuakeMap) {
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorDepth float64 = 64.0 * scale

	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			actor := rtlmap.ActorGrid[y][x]
			enemy := actor.Enemy
			if enemy != nil {
				entityName := enemy.ConversionInfo.EntityName(&actor, dusk)
				if entityName == "" {
					continue
				}
				entity := qm.SpawnEntity(entityName, 0)
				AddDefaultEntityKeys(entity, &actor)
				entity.OriginX = (float64(x) + 0.5) * gridSizeX
				entity.OriginY = (float64(y) + 0.5) * -gridSizeY

				var angle float64
				switch enemy.Direction {
				case DIR_East:
					angle = 0.0
				case DIR_North:
					angle = 90.0
				case DIR_West:
					angle = 180.0
				case DIR_South:
					angle = 270.0
				default:
					log.Panicf("Unknown enemy direction: %d", int(enemy.Direction))
				}
				entity.AdditionalKeys["angle"] = fmt.Sprintf("%.02f", angle)

				switch enemy.Difficulty {
				case DifficultyEasy:
					entity.SpawnFlags = quakemap.SPAWNFLAG_NotOnHard
				}

				// TODO: Z axis placement needs to be cleaned up. Lots
				// of redundant code that's also in item placement
				switch {
				case actor.InfoValue&0xff00 == 0xb000:
					entity.OriginZ = floorDepth + (gridSizeZ / 2.0) + rtlmap.ZOffset(actor.InfoValue, scale)
				case actor.Type == WALL_Platform:
					// rt_door.c:243
					var itemZOffset float64
					switch actor.InfoValue {
					case 1, 8, 9:
						itemZOffset = float64(rtlmap.FloorHeight()-1) * gridSizeZ
					case 4, 7:
						itemZOffset = 0.0
					case 5, 6:
						itemZOffset = gridSizeZ
					}
					entity.OriginZ = floorDepth + itemZOffset + (gridSizeZ / 2.0)
				default:
					entity.OriginZ = floorDepth + (gridSizeZ / 2.0)
				}
			}
		}
	}
}

func ConvertRTLMapToQuakeMapFile(rtlmap *RTLMapData, textureWad string, scale float64, dusk bool, additionalWads []string, fgdFile string) *quakemap.QuakeMap {

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
	additionalWads = append(additionalWads, textureWad)
	qm.Wads = additionalWads
	if fgdFile != "" {
		qm.WorldSpawn.AdditionalKeys["_tb_def"] = fmt.Sprintf("external:%s", fgdFile)
	}

	floorBrush := quakemap.BasicCuboid(
		0, 0, 0,
		floorWidth, -floorLength, floorDepth,
		rtlmap.FloorTexture(),
		scale, false)
	qm.WorldSpawn.AddBrush(floorBrush)

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
		qm.WorldSpawn.AddBrush(ceilBrush)
	}

	// spawn walls, items, static entities
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			wallInfo := rtlmap.ActorGrid[y][x]
			itemInfo := rtlmap.ActorGrid[y][x].Item

			switch wallInfo.Type {
			case WALL_Regular, WALL_Elevator:
				CreateRegularWall(rtlmap, x, y, scale, qm)
			case WALL_ThinWall:
				CreateThinWall(rtlmap, x, y, scale, dusk, qm)
			case WALL_AnimatedWall:
				CreateRegularWall(rtlmap, x, y, scale, qm)
			case WALL_Platform:
				CreatePlatform(rtlmap, x, y, scale, qm)
			case WALL_MaskedWall:
				CreateMaskedWall(rtlmap, x, y, scale, dusk, qm)
			case SPR_GAD:
				CreateGAD(rtlmap, &wallInfo, scale, dusk, qm)
			}

			if itemInfo != nil {
				if itemInfo.AddCallback != nil {
					itemInfo.AddCallback(x, y, gridSizeX, gridSizeY, gridSizeZ, itemInfo, rtlmap, qm, dusk)
				} else {
					entityName := itemInfo.QuakeEntityName
					if dusk {
						entityName = itemInfo.DuskEntityName
					}

					if entityName == "" {
						continue
					}

					entity := qm.SpawnEntity(entityName, 0)
					entity.OriginX = float64(x)*gridSizeX + (gridSizeX / 2.0)
					entity.OriginY = float64(y)*-gridSizeY - (gridSizeY / 2.0)
					switch {
					case wallInfo.InfoValue&0xff00 == 0xb000:
						entity.OriginZ = floorDepth + (gridSizeZ / 2.0) + rtlmap.ZOffset(wallInfo.InfoValue, scale)
					case wallInfo.Type == WALL_Platform:
						// rt_door.c:243
						var itemZOffset float64
						switch wallInfo.InfoValue {
						case 1, 8, 9:
							itemZOffset = float64(rtlmap.FloorHeight()-1) * gridSizeZ
						case 4, 7:
							itemZOffset = 0.0
						case 5, 6:
							itemZOffset = -gridSizeZ
						}
						entity.OriginZ = floorDepth + itemZOffset
						if !itemInfo.PlaceOnFloor {
							// center it just above the platform
							entity.OriginZ += (gridSizeZ / 2.0)
						}
					case wallInfo.InfoValue == 11, wallInfo.InfoValue == 12:
						entity.OriginZ = floorDepth - 65.0 - float64(wallInfo.InfoValue-11)
					case itemInfo.PlaceOnFloor == true:
						if dusk {
							entity.OriginZ = floorDepth + (float64(itemInfo.DuskHeight) / 2.0) + itemInfo.DuskZOffset
						} else {
							entity.OriginZ = floorDepth + (float64(itemInfo.QuakeHeight) / 2.0) + itemInfo.QuakeZOffset
						}
					default:
						entity.OriginZ = floorDepth + (gridSizeZ / 2.0)
					}
				}
			}

		}
	}

	// spawn touchplate triggers
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			actor := &rtlmap.ActorGrid[y][x]
			if len(actor.MapTriggers) > 0 {
				log.Printf("Creating triggers at (%d,%d)", actor.X, actor.Y)
				CreateTrigger(rtlmap, actor, scale, qm)
			}
		}
	}

	CreateDoorEntities(rtlmap, scale, dusk, qm)
	LinkElevators(rtlmap, textureWad, floorDepth, gridSizeX, gridSizeY, gridSizeZ, scale, dusk, qm)
	AddExitPoints(rtlmap, scale, dusk, qm)
	AddEnemies(rtlmap, scale, dusk, qm)

	// 2. TODO: clip brushes around floor extending height
	return qm
}
