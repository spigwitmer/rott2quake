package rtl

import (
	"fmt"
	"gitlab.com/camtap/lumps/pkg/quakemap"
)

var (
	gridSizeX float64 = 64
	gridSizeY float64 = 64
	gridSizeZ float64 = 64
	exitLumps         = []string{
		"EXIT",
		"ENTRANCE",
		"EXITARCH",
		"EXITARCA",
		"ENTRARCH",
	}
)

func wallTileToTextureName(wallInfo *WallInfo) string {
	// TODO: correlate with WALLSTRT and EXITSTRT lumps in WAD
	tileId := wallInfo.Tile
	if tileId >= 1 && tileId <= 32 {
		return fmt.Sprintf("WALL%d", tileId)
	} else if tileId >= 36 && tileId <= 45 {
		return fmt.Sprintf("WALL%d", tileId-3)
	} else if tileId == 46 {
		return "WALL73"
	} else if tileId == 47 || tileId == 48 {
		return exitLumps[tileId-47]
	} else if tileId >= 49 && tileId <= 71 {
		return fmt.Sprintf("WALL%d", tileId-8)
	} else if tileId >= 72 && tileId <= 79 {
		return fmt.Sprintf("ELEV%d", tileId-71)
	} else if tileId >= 80 && tileId <= 89 {
		return fmt.Sprintf("WALL%d", tileId-16)
	} else {
		panic(fmt.Sprintf("Bad tile ID for wall: 0x%x", tileId))
	}
}

func ConvertRTLMapToQuakeMapFile(rtlmap *RTLMapData, textureWad string) *quakemap.QuakeMap {

	// worldspawn:
	// 1. build 128x128 floor
	var floorLength float64 = gridSizeX * 128
	var floorWidth float64 = gridSizeY * 128
	var floorDepth float64 = 64

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
		1)
	qm.WorldSpawn.Brushes = []quakemap.Brush{floorBrush}

	// place static walls
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			wallInfo := rtlmap.CookedWallGrid[i][j]
			if wallInfo.Type != WALL_None {
				wallColumn := quakemap.BasicCuboid(
					float64(i)*gridSizeX,   // x1
					float64(j)*gridSizeY,   // y1
					floorDepth,             // z1
					float64(i+1)*gridSizeX, // x2
					float64(j+1)*gridSizeY, // y2
					floorDepth+float64(rtlmap.FloorHeight())*gridSizeZ, // z2
					wallTileToTextureName(&wallInfo),
					1) // scale

				qm.WorldSpawn.Brushes = append(qm.WorldSpawn.Brushes, wallColumn)
			}
		}
	}

	// 2. TODO: clip brushes around floor extending height
	return qm
}
