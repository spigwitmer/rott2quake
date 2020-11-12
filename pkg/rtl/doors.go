package rtl

import (
    "fmt"
)

type DoorTexInfo struct {
    BaseTexture string
    SideTexture string
    AltTexture string
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

type Door struct {
    Lock DoorLock
    TriggerX int
    TriggerY int
    Direction WallDirection
    Tiles []*WallInfo
}

// texture ID --> Door texture names
var DoorTextures = map[uint16]DoorTexInfo{
    0: DoorTexInfo{"RAMDOOR1", "SIDE8", "ABOVEW3"},
    1: DoorTexInfo{"DOOR2", "SIDE8", "ABOVEW3"},
    2: DoorTexInfo{"TRIDOOR1", "SIDE8", "ABOVEW3"},
    3: DoorTexInfo{"TRIDOOR1", "SIDE8", "ABOVEW3"},
    8: DoorTexInfo{"RAMDOOR1", "SIDE8", "ABOVEW3"},
    9: DoorTexInfo{"DOOR2", "SIDE8", "ABOVEW3"},
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
    if tileId >= 33 && tileId <= 35 {
        doorId = tileId - 33 + 15
    } else if tileId >= 90 && tileId <= 93 {
        doorId = tileId - 90
    } else if tileId >= 98 && tileId <= 104 {
        doorId = tileId - 90
    } else if tileId >= 94 && tileId <= 97 {
        doorId = tileId - 86
    } else if tileId >= 154 && tileId <= 156 {
        doorId = tileId - 154 + 18
    }
    if ok, texInfo := DoorTextures[doorId]; ok {
        return &texInfo
    } else {
        return nil
    }
}

func (r *RTLMapData) GetDoors() []Door {
    mapTileToDoor := make(map[string]*Door)
    var doors []Door
    for x := 0; x < 128; x++ {
        for y := 0; y < 128; y++ {
            mapKey := fmt.Sprintf("%d%d", x, y)
            if _, ok := mapTileToDoorInfo[mapKey]; ok {
                // tile was already processed
                continue
            }
            if texInfo := GetDoorTextures(r.CookedWallGrid[x][y]); texInfo != nil {
                // door placed here, find neighboring door tiles
                var newDoor Door
                newDoor.Tiles = append(newDoor.Tiles, r.CookedWallGrid[x][y])
                mapTileToDoorInfo[mapKey] = &newDoor

                // TODO: key sprites?
                // TODO: touchplate trigger locations?

                // find adjacent door tiles north of it
                if y > 0 && addTexInfo := GetDoorTextures(r.CookedWallGrid[x][y-1]); {
                    adjacentKey := fmt.Sprintf("%d%d", x, y-1)
                    if _, ok := mapTileToDoorInfo[adjacentKey]; ok {
                        continue
                    }
                    newDoor.Tiles = append(newDoor.Tiles, r.CookedWallGrid[x][y-1])
                    mapTileToDoorInfo[adjacentKey] = &newDoor
                    newDoor.Direction = WALLDIR_NorthSouth

                    for ay := y-2; ay >= 0 && addTexInfo != nil; ay-- {
                        adjacentKey := fmt.Sprintf("%d%d", x, ay)
                        if _, ok := mapTileToDoorInfo[adjacentKey]; ok {
                            continue
                        }
                        addTexInfo = GetDoorTextures(r.CookedWallGrid[x][ay])
                        if addTexInfo != nil {
                            newDoor.Tiles = append(newDoor.Tiles, r.CookedWallGrid[x][ay])
                            mapTileToDoorInfo[adjacentKey] = &newDoor
                        }
                    }
                }
                // south of it
                if y < 127 && addTexInfo := GetDoorTextures(r.CookedWallGrid[x][y+1]); {
                    adjacentKey := fmt.Sprintf("%d%d", x, y+1)
                    if _, ok := mapTileToDoorInfo[adjacentKey]; ok {
                        continue
                    }
                    newDoor.Tiles = append(newDoor.Tiles, r.CookedWallGrid[x][y+1])
                    mapTileToDoorInfo[adjacentKey] = &newDoor
                    newDoor.Direction = WALLDIR_NorthSouth

                    for ay := y+2; ay < 128 && addTexInfo != nil; ay++ {
                        adjacentKey := fmt.Sprintf("%d%d", x, ay)
                        if _, ok := mapTileToDoorInfo[adjacentKey]; ok {
                            continue
                        }
                        addTexInfo = GetDoorTextures(r.CookedWallGrid[x][ay])
                        if addTexInfo != nil {
                            newDoor.Tiles = append(newDoor.Tiles, r.CookedWallGrid[x][ay])
                            mapTileToDoorInfo[adjacentKey] = &newDoor
                        }
                    }
                }
                // west of it
                if x > 0 && addTexInfo := GetDoorTextures(r.CookedWallGrid[x-1][y]); {
                    adjacentKey := fmt.Sprintf("%d%d", x-1, y)
                    if _, ok := mapTileToDoorInfo[adjacentKey]; ok {
                        continue
                    }
                    newDoor.Tiles = append(newDoor.Tiles, r.CookedWallGrid[x-1][y])
                    mapTileToDoorInfo[adjacentKey] = &newDoor
                    newDoor.Direction = WALLDIR_EastWest

                    for ax := x-2; ax >= 0 && addTexInfo != nil; ax-- {
                        adjacentKey := fmt.Sprintf("%d%d", ax, y)
                        if _, ok := mapTileToDoorInfo[adjacentKey]; ok {
                            continue
                        }
                        addTexInfo = GetDoorTextures(r.CookedWallGrid[ax][y])
                        if addTexInfo != nil {
                            newDoor.Tiles = append(newDoor.Tiles, r.CookedWallGrid[ax][y])
                            mapTileToDoorInfo[adjacentKey] = &newDoor
                        }
                    }
                }
                // east of it
                if x < 127 && addTexInfo := GetDoorTextures(r.CookedWallGrid[x+1][y]); {
                    adjacentKey := fmt.Sprintf("%d%d", x+1, y)
                    if _, ok := mapTileToDoorInfo[adjacentKey]; ok {
                        continue
                    }
                    newDoor.Tiles = append(newDoor.Tiles, r.CookedWallGrid[x+1][y])
                    mapTileToDoorInfo[adjacentKey] = &newDoor
                    newDoor.Direction = WALLDIR_EastWest

                    for ax := x+2; ax < 128 && addTexInfo != nil; ax++ {
                        adjacentKey := fmt.Sprintf("%d%d", ax, y)
                        if _, ok := mapTileToDoorInfo[adjacentKey]; ok {
                            continue
                        }
                        addTexInfo = GetDoorTextures(r.CookedWallGrid[ax][y])
                        if addTexInfo != nil {
                            newDoor.Tiles = append(newDoor.Tiles, r.CookedWallGrid[ax][y])
                            mapTileToDoorInfo[adjacentKey] = &newDoor
                        }
                    }
                }
            }
        }
    }

    return doors
}
