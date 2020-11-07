package rtl

import (
	"fmt"
	"gitlab.com/camtap/lumps/pkg/quakemap"
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
	if w.Flags&(MWF_Shootable|MWF_BlockingChanges) == (MWF_Shootable | MWF_BlockingChanges) {
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

func ConvertRTLMapToQuakeMapFile(rtlmap *RTLMapData, textureWad string, scale float64, dusk bool) *quakemap.QuakeMap {

	// worldspawn:
	// 1. build 128x128 floor
	var gridSizeX float64 = 64.0 * scale
	var gridSizeY float64 = 64.0 * scale
	var gridSizeZ float64 = 64.0 * scale
	var floorLength float64 = gridSizeX * 128
	var floorWidth float64 = gridSizeY * 128
	var floorDepth float64 = 64 * scale

	var playerStartX float64 = float64(rtlmap.SpawnX)*gridSizeX + (gridSizeX / 2)
	var playerStartY float64 = float64(rtlmap.SpawnY)*gridSizeY + (gridSizeY / 2)
	var playerAngle float64
	switch rtlmap.SpawnDirection {
	case 0: // up
		playerAngle = 180
	case 1: // right
		playerAngle = 90
	case 2: // down
		playerAngle = 0
	case 3: // left
		playerAngle = 270
	}

	qm := quakemap.NewQuakeMap(playerStartX, playerStartY, floorDepth+32)
	qm.InfoPlayerStart.Angle = playerAngle
	qm.Wad = textureWad

	floorBrush := quakemap.BasicCuboid(
		0, 0, 0,
		floorWidth, floorLength, floorDepth,
		rtlmap.FloorTexture(),
		scale)
	qm.WorldSpawn.Brushes = []quakemap.Brush{floorBrush}

	// add ceiling if declared
	ceilTexture := rtlmap.CeilingTexture()
	if ceilTexture != "" {
		ceilz1 := floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
		ceilz2 := floorDepth + float64(rtlmap.FloorHeight()+1)*gridSizeZ
		ceilBrush := quakemap.BasicCuboid(
			0, 0, ceilz1,
			floorWidth, floorLength, ceilz2,
			ceilTexture,
			scale)
		qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, ceilBrush)
	}

	floorHeight := rtlmap.FloorHeight()

	// place walls
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			wallInfo := rtlmap.CookedWallGrid[i][j]
			texName := wallInfo.WallTileToTextureName(false)
			if wallInfo.Type == WALL_Regular || wallInfo.Type == WALL_AnimatedWall || wallInfo.Type == WALL_Elevator {
				var x1, y1, x2, y2 float64
				infoVal := rtlmap.InfoPlane[i][j]
				spriteVal := rtlmap.SpritePlane[i][j]
				wallDirection := rtlmap.ThinWallDirection(i, j)
				if infoVal == 1 || (infoVal >= 4 && infoVal <= 9) {
					if wallDirection == WALLDIR_NorthSouth {
						x1 = float64(i)*gridSizeX + (gridSizeX / 2) - 2
						x2 = float64(i)*gridSizeX + (gridSizeX / 2) + 2
						y1 = float64(j) * gridSizeY
						y2 = float64(j+1) * gridSizeY
					} else {
						x1 = float64(i) * gridSizeX
						x2 = float64(i+1) * gridSizeX
						y1 = float64(j)*gridSizeY + (gridSizeY / 2) - 2
						y2 = float64(j)*gridSizeY + (gridSizeY / 2) + 2
					}
					switch infoVal {
					case 1:
						// thin wall, above passable
						var z1 float64 = floorDepth
						var z2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
						wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
							texName, scale)
						qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
					case 4:
						// thin wall, above only
						var z1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
						var z2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
						wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
							texName, scale)
						qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
					case 5:
						// thin wall, below only
						var z1 float64 = floorDepth
						var z2 float64 = floorDepth + gridSizeZ
						wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
							texName, scale)
						qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
					case 6:
						// thin wall, middle passable
						var bottomz1 float64 = floorDepth
						var bottomz2 float64 = floorDepth + gridSizeZ
						var topz1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
						var topz2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
						wallColumn1 := quakemap.BasicCuboid(x1, y1, bottomz1, x2, y2, bottomz2,
							texName, scale)
						wallColumn2 := quakemap.BasicCuboid(x1, y1, topz1, x2, y2, topz2,
							texName, scale)
						qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn1)
						qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn2)
					case 7:
						// thin wall, everything but below
						var z1 float64 = floorDepth + gridSizeZ
						var z2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
						wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
							texName, scale)
						qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
					case 8:
						// thin wall, middle only
						var z1 float64 = floorDepth + gridSizeZ
						var z2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
						wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
							texName, scale)
						qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
					case 9:
						// thin wall, everything but above
						var z1 float64 = floorDepth
						var z2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
						wallColumn := quakemap.BasicCuboid(x1, y1, z1, x2, y2, z2,
							texName, scale)
						qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
					}
				} else {
					// plain ol' column
					wallColumn := quakemap.BasicCuboid(
						float64(i)*gridSizeX,   // x1
						float64(j)*gridSizeY,   // y1
						floorDepth,             // z1
						float64(i+1)*gridSizeX, // x2
						float64(j+1)*gridSizeY, // y2
						floorDepth+float64(rtlmap.FloorHeight())*gridSizeZ, // z2
						texName,
						scale) // scale

					// make static walls part of the worldspawn,
					// everything else a separate entity
					if spriteVal == 0 && infoVal == 0 && wallInfo.Type == WALL_Regular {
						qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
					} else {
						entity := quakemap.NewEntity(0, "func_wall", qm)
						entity.Brushes = []quakemap.Brush{wallColumn}
						qm.Entities = append(qm.Entities, entity)
					}
				}

			} else if wallInfo.Type == WALL_Platform {
				// platforms are supposed to work like masked walls,
				// however just implement that as full walls since
				// they're made to allow the player walk across
				if platformInfo, ok := Platforms[wallInfo.PlatformID]; ok {
					x1 := float64(i) * gridSizeX + 1
					y1 := float64(j) * gridSizeY + 1
					x2 := float64(i+1) * gridSizeX - 1
					y2 := float64(j+1) * gridSizeY - 1

					// above as separate entity
					// NOTE: don't render tops and bottoms of platforms
					// if they're passable, they look nasty
					if platformInfo.Above != "" && platformInfo.Flags&MWF_AbovePassable == 0 && floorHeight > 1 {
						var abovez1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
						var abovez2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
						aboveClassName := "func_detail"
						aboveColumn := quakemap.BasicCuboid(x1, y1, abovez1, x2, y2, abovez2,
							"{"+platformInfo.Above,
							scale)
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
							scale)
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
							scale)
						bottomEntity := quakemap.NewEntity(0, className, qm)
						bottomEntity.Brushes = append(bottomEntity.Brushes, column)
						qm.Entities = append(qm.Entities, bottomEntity)
					}
				}
			} else if wallInfo.Type == WALL_MaskedWall {
				// masked walls have adjacent sides, a thin wall in the
				// middle, and the bottom may be passable
				if maskedWallInfo, ok := MaskedWalls[wallInfo.Tile]; ok {
					wallDirection := rtlmap.ThinWallDirection(i, j)
					var x1, y1, x2, y2 float64

					if wallDirection == WALLDIR_NorthSouth {
						x1 = float64(i)*gridSizeX + (gridSizeX / 2)
						x2 = float64(i)*gridSizeX + (gridSizeX / 2) + 1
						y1 = float64(j) * gridSizeY
						y2 = float64(j+1) * gridSizeY
					} else {
						x1 = float64(i) * gridSizeX
						x2 = float64(i+1) * gridSizeX
						y1 = float64(j)*gridSizeY + (gridSizeY / 2)
						y2 = float64(j)*gridSizeY + (gridSizeY / 2) + 1
					}

					// above as separate entity
					if maskedWallInfo.Above != "" && floorHeight > 1 {
						var abovez1 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
						var abovez2 float64 = floorDepth + float64(rtlmap.FloorHeight())*gridSizeZ
						aboveClassName := ClassNameForMaskedWall(&maskedWallInfo, "above")
						aboveColumn := quakemap.BasicCuboid(x1, y1, abovez1, x2, y2, abovez2,
							"{"+maskedWallInfo.Above,
							scale)
						aboveEntity := quakemap.NewEntity(0, aboveClassName, qm)
						aboveEntity.Brushes = append(aboveEntity.Brushes, aboveColumn)
						qm.Entities = append(qm.Entities, aboveEntity)
					}

					// middle
					if maskedWallInfo.Middle != "" && floorHeight > 2 {
						var middlez1 float64 = floorDepth + gridSizeZ
						var middlez2 float64 = floorDepth + float64(rtlmap.FloorHeight()-1)*gridSizeZ
						middleClassName := ClassNameForMaskedWall(&maskedWallInfo, "middle")
						mwColumn := quakemap.BasicCuboid(x1, y1, middlez1, x2, y2, middlez2,
							"{"+maskedWallInfo.Middle,
							scale)
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
							scale)
						bottomEntity := quakemap.NewEntity(0, className, qm)
						bottomEntity.Brushes = append(bottomEntity.Brushes, column)
						qm.Entities = append(qm.Entities, bottomEntity)
					}

					// TODO: sides

				} else {
					panic(fmt.Sprintf("Masked wall at %d,%d has non-existent ID (%d)", i, j, wallInfo.MaskedWallID))
				}
			}
		}
	}

	// place items
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			itemInfo := rtlmap.CookedSpriteGrid[i][j].Item
			if itemInfo != nil {
				if itemInfo.AddCallback != nil {
					itemInfo.AddCallback(i, j, gridSizeX, gridSizeY, gridSizeZ, itemInfo, rtlmap, qm, dusk)
				} else {
					entityName := itemInfo.QuakeEntityName
					if dusk {
						entityName = itemInfo.DuskEntityName
					}
					entity := quakemap.NewEntity(0, entityName, qm)
					entity.OriginX = float64(i)*gridSizeX + (gridSizeX / 2)
					entity.OriginY = float64(j)*gridSizeY + (gridSizeY / 2)
					entity.OriginZ = floorDepth + (gridSizeZ / 2)
					qm.Entities = append(qm.Entities, entity)
				}
			}
		}
	}

	// 2. TODO: clip brushes around floor extending height
	return qm
}
