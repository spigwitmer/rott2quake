package rtl

var (
	// any tile value in the first plane above this number is part of an area
	areaTileMin uint16 = 107
)

func CalculateAreas(rtlmap *RTLMapData) {
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			wallVal := rtlmap.WallPlane[i][j]
			if wallVal >= areaTileMin {
				rtlmap.CookedWallGrid[i][j].AreaID = int(wallVal) - 107
			}
		}
	}
}
