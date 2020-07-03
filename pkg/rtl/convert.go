package rtl

import (
	"gitlab.com/camtap/lumps/pkg/quakemap"
)

var (
	gridSizeX float64 = 64
	gridSizeY float64 = 64
)

func ConvertRTLMapToQuakeMapFile(rtlmap *RTLMapData, textureWad string) *quakemap.QuakeMap {

	// worldspawn:
	// 1. build 128x128 floor
	var floorLength float64 = gridSizeX * 128
	var floorWidth float64 = gridSizeY * 128
	var floorDepth float64 = 64
	var floorBrush quakemap.Brush

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

	// south
	floorBrush.AddPlane(
		0.0, 0.0, 0.0, // p1
		0.0, 0.0, 1.0, // p2
		1.0, 0.0, 0.0, // p3
		rtlmap.FloorTexture(), // texture
		0, 0,                  // offset
		0,    // rotation
		1, 1) // scale

	// north
	floorBrush.AddPlane(
		0.0, floorLength, 0.0, // p1
		1.0, floorLength, 0.0, // p2
		0.0, floorLength, 1.0, // p3
		rtlmap.FloorTexture(), // texture
		0, 0,                  // offset
		0,    // rotation
		1, 1) // scale

	// east
	floorBrush.AddPlane(
		floorWidth, 0.0, 0.0, // p1
		floorWidth, 0.0, 1.0, // p2
		floorWidth, 1.0, 0.0, // p3
		rtlmap.FloorTexture(), // texture
		0, 0,                  // offset
		0,    // rotation
		1, 1) // scale

	// west
	floorBrush.AddPlane(
		0.0, 0.0, 0.0, // p1
		0.0, 1.0, 0.0, // p2
		0.0, 0.0, 1.0, // p3
		rtlmap.FloorTexture(), // texture
		0, 0,                  // offset
		0,    // rotation
		1, 1) // scale

	// top
	floorBrush.AddPlane(
		0.0, 0.0, floorDepth, // p1
		0.0, 1.0, floorDepth, // p2
		1.0, 0.0, floorDepth, // p3
		rtlmap.FloorTexture(), // texture
		0, 0,                  // offset
		0,    // rotation
		1, 1) // scale

	// bottom
	floorBrush.AddPlane(
		0.0, 0.0, 0.0, // p1
		1.0, 0.0, 0.0, // p2
		0.0, 1.0, 0.0, // p3
		rtlmap.FloorTexture(), // texture
		0, 0,                  // offset
		0,    // rotation
		1, 1) // scale

	qm.WorldSpawn.Brushes = []quakemap.Brush{floorBrush}

	// 2. clip brushes around floor extending height
	return qm
}
