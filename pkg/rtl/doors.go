package rtl

import (
	"fmt"
)

type DoorTexInfo struct {
	BaseTexture string
	SideTexture string
	AltTexture  string
}

type DoorLock int

const (
	LOCK_Unlocked DoorLock = iota
	LOCK_GoldKey
	LOCK_SilverKey
	LOCK_IronKey
	LOCK_OscuroKey
	LOCK_Trigger // Touchplate trigger
)

func (d DoorLock) KeyName() string {
	switch d {
	case LOCK_GoldKey:
		return "Gold"
	case LOCK_SilverKey:
		return "Silver"
	case LOCK_IronKey:
		return "Iron"
	case LOCK_OscuroKey:
		return "Oscuro"
	default:
		return "(no key)"
	}
}

type Door struct {
	Lock      DoorLock
	TriggerX  int
	TriggerY  int
	Direction WallDirection
	Tiles     []ActorInfo
}

// texture ID --> Door texture names
var DoorTextures = map[uint16]DoorTexInfo{
	0:  DoorTexInfo{"RAMDOOR1", "SIDE8", "ABOVEW3"},
	1:  DoorTexInfo{"DOOR2", "SIDE8", "ABOVEW3"},
	2:  DoorTexInfo{"TRIDOOR1", "SIDE8", "ABOVEW3"},
	3:  DoorTexInfo{"TRIDOOR1", "SIDE8", "ABOVEW3"},
	8:  DoorTexInfo{"RAMDOOR1", "SIDE8", "ABOVEW3"},
	9:  DoorTexInfo{"DOOR2", "SIDE8", "ABOVEW3"},
	10: DoorTexInfo{"SDOOR4", "SIDE8", "ABOVEW3"},
	11: DoorTexInfo{"SDOOR4", "SIDE8", "ABOVEW3"},
	12: DoorTexInfo{"EDOOR", "SIDE8", "ABOVEW3"},
	13: DoorTexInfo{"TRIDOOR1", "SIDE8", "ABOVEW3"},
	14: DoorTexInfo{"SDOOR4", "SIDE8", "ABOVEW3"},
	15: DoorTexInfo{"SNDOOR", "SIDE16", "ABOVEW16"},
	16: DoorTexInfo{"SNADOOR", "SIDE16", "ABOVEW16"},
	17: DoorTexInfo{"SNKDOOR", "SIDE16", "ABOVEW16"},
	18: DoorTexInfo{"TNDOOR", "SIDE17", "ABOVEW17"},
	19: DoorTexInfo{"TNADOOR", "SIDE17", "ABOVEW17"},
	20: DoorTexInfo{"TNKDOOR", "SIDE17", "ABOVEW17"},
}

func GetDoorTextures(tileID uint16) *DoorTexInfo {
	var doorId uint16 = 99
	if tileID >= 33 && tileID <= 35 {
		doorId = tileID - 33 + 15
	} else if tileID >= 90 && tileID <= 93 {
		doorId = tileID - 90
	} else if tileID >= 98 && tileID <= 104 {
		doorId = tileID - 90
	} else if tileID >= 94 && tileID <= 97 {
		doorId = tileID - 86
	} else if tileID >= 154 && tileID <= 156 {
		doorId = tileID - 154 + 18
	}
	if texInfo, ok := DoorTextures[doorId]; ok {
		return &texInfo
	} else {
		return nil
	}
}

func (r *RTLMapData) GetDoors() []Door {
	mapTileToDoor := make(map[string]*Door)
	var doors []Door
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			mapKey := fmt.Sprintf("%d%d", x, y)
			if _, ok := mapTileToDoor[mapKey]; ok {
				// tile was already processed
				continue
			}
			if r.ActorGrid[y][x].Type == WALL_Door {
				// door placed here, find neighboring door tiles
				var newDoor Door
				newDoor.Lock = LOCK_Unlocked
				newDoor.Tiles = append(newDoor.Tiles, r.ActorGrid[y][x])
				mapTileToDoor[mapKey] = &newDoor
				directionKnown := false

				// add keys (sprite val takes precedence)
				if r.WallPlane[y][x] >= 94 && r.WallPlane[y][x] <= 97 {
					newDoor.Lock = DoorLock(int(r.WallPlane[y][x] - 93))
				}
				if r.SpritePlane[y][x] >= 0x1d && r.SpritePlane[y][x] <= 0x20 {
					newDoor.Lock = DoorLock(int(r.SpritePlane[y][x] - 0x1c))
				}

				// TODO: touchplate trigger locations?

				// find adjacent door tiles north of it
				if y > 0 && r.ActorGrid[y-1][x].Type == WALL_Door {
					addTexInfo := GetDoorTextures(r.ActorGrid[y-1][x].Tile)
					adjacentKey := fmt.Sprintf("%d%d", x, y-1)
					if _, ok := mapTileToDoor[adjacentKey]; ok {
						continue
					}
					newDoor.Tiles = append(newDoor.Tiles, r.ActorGrid[y-1][x])
					mapTileToDoor[adjacentKey] = &newDoor
					newDoor.Direction = WALLDIR_NorthSouth
					directionKnown = true

					for ay := y - 2; ay >= 0 && addTexInfo != nil; ay-- {
						adjacentKey := fmt.Sprintf("%d%d", x, ay)
						if _, ok := mapTileToDoor[adjacentKey]; ok {
							break
						}
						if r.ActorGrid[ay][x].Type != WALL_Door {
							break
						}
						addTexInfo = GetDoorTextures(r.ActorGrid[ay][x].Tile)
						newDoor.Tiles = append(newDoor.Tiles, r.ActorGrid[ay][x])
						mapTileToDoor[adjacentKey] = &newDoor
					}
				}
				// south of it
				if y < 127 && r.ActorGrid[y+1][x].Type == WALL_Door {
					addTexInfo := GetDoorTextures(r.ActorGrid[y+1][x].Tile)
					adjacentKey := fmt.Sprintf("%d%d", x, y+1)
					if _, ok := mapTileToDoor[adjacentKey]; ok {
						continue
					}
					newDoor.Tiles = append(newDoor.Tiles, r.ActorGrid[y+1][x])
					mapTileToDoor[adjacentKey] = &newDoor
					newDoor.Direction = WALLDIR_NorthSouth
					directionKnown = true

					for ay := y + 2; ay < 128 && addTexInfo != nil; ay++ {
						adjacentKey := fmt.Sprintf("%d%d", x, ay)
						if _, ok := mapTileToDoor[adjacentKey]; ok {
							break
						}
						if r.ActorGrid[ay][x].Type != WALL_Door {
							break
						}
						addTexInfo = GetDoorTextures(r.ActorGrid[ay][x].Tile)
						newDoor.Tiles = append(newDoor.Tiles, r.ActorGrid[ay][x])
						mapTileToDoor[adjacentKey] = &newDoor
					}
				}
				// west of it
				if x > 0 && r.ActorGrid[y][x-1].Type == WALL_Door {
					addTexInfo := GetDoorTextures(r.ActorGrid[y][x-1].Tile)
					adjacentKey := fmt.Sprintf("%d%d", x-1, y)
					if _, ok := mapTileToDoor[adjacentKey]; ok {
						continue
					}
					newDoor.Tiles = append(newDoor.Tiles, r.ActorGrid[y][x-1])
					mapTileToDoor[adjacentKey] = &newDoor
					newDoor.Direction = WALLDIR_EastWest
					directionKnown = true

					for ax := x - 2; ax >= 0 && addTexInfo != nil; ax-- {
						adjacentKey := fmt.Sprintf("%d%d", ax, y)
						if _, ok := mapTileToDoor[adjacentKey]; ok {
							break
						}
						if r.ActorGrid[y][ax].Type != WALL_Door {
							break
						}
						addTexInfo = GetDoorTextures(r.ActorGrid[y][ax].Tile)
						newDoor.Tiles = append(newDoor.Tiles, r.ActorGrid[y][ax])
						mapTileToDoor[adjacentKey] = &newDoor
					}
				}
				// east of it
				if x < 127 && r.ActorGrid[y][x+1].Type == WALL_Door {
					addTexInfo := GetDoorTextures(r.ActorGrid[y][x+1].Tile)
					adjacentKey := fmt.Sprintf("%d%d", x+1, y)
					if _, ok := mapTileToDoor[adjacentKey]; ok {
						continue
					}
					newDoor.Tiles = append(newDoor.Tiles, r.ActorGrid[y][x+1])
					mapTileToDoor[adjacentKey] = &newDoor
					newDoor.Direction = WALLDIR_EastWest
					directionKnown = true

					for ax := x + 2; ax < 128 && addTexInfo != nil; ax++ {
						adjacentKey := fmt.Sprintf("%d%d", ax, y)
						if _, ok := mapTileToDoor[adjacentKey]; ok {
							break
						}
						if r.ActorGrid[y][ax].Type != WALL_Door {
							break
						}
						addTexInfo = GetDoorTextures(r.ActorGrid[y][ax].Tile)
						newDoor.Tiles = append(newDoor.Tiles, r.ActorGrid[y][ax])
						mapTileToDoor[adjacentKey] = &newDoor
					}
				}

				if !directionKnown {
					newDoor.Direction, _, _ = r.ThinWallDirection(x, y)
				}

				doors = append(doors, newDoor)
			}
		}
	}

	return doors
}
