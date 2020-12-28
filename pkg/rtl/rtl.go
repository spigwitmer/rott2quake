package rtl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"html/template"
	"io"
	"log"
	"strings"
)

var (
	rtlMagic = [4]byte{'R', 'T', 'L', '\x00'}
	// any tile value in the first plane above this number is part of an area
	AreaTileMin uint16 = 107
	NumAreas    uint16 = 47
)

type RTLHeader struct {
	Signature [4]byte
	Version   uint32
}

type RTLMapHeader struct {
	Used              uint32
	CRC               [4]byte
	RLEWTag           uint32
	MapSpecials       uint32
	WallPlaneOffset   uint32
	SpritePlaneOffset uint32
	InfoPlaneOffset   uint32
	WallPlaneLength   uint32
	SpritePlaneLength uint32
	InfoPlaneLength   uint32
	Name              [24]byte
}

// "Actors" comprise of pretty much everything: walls, switches,
// enemies, etc.
type ActorType int

func (a *ActorType) String() string {
	switch *a {
	case ACTOR_None:
		return "ACTOR_None"
	case WALL_Regular:
		return "WALL_Regular"
	case WALL_ThinWall:
		return "WALL_ThinWall"
	case WALL_Elevator:
		return "WALL_Elevator"
	case WALL_AnimatedWall:
		return "WALL_AnimatedWall"
	case WALL_MaskedWall:
		return "WALL_MaskedWall"
	case WALL_Platform:
		return "WALL_Platform"
	case WALL_Window:
		return "WALL_Window"
	case WALL_PushWall:
		return "WALL_PushWall"
	case WALL_Switch:
		return "WALL_Switch"
	case WALL_Door:
		return "WALL_Door"
	case SPR_GAD:
		return "SPR_GAD"
	case SPR_Static:
		return "SPR_Static"
	default:
		return "unknown"
	}
}

const (
	ACTOR_None ActorType = iota
	WALL_Regular
	WALL_ThinWall
	WALL_Elevator
	WALL_AnimatedWall
	WALL_MaskedWall
	WALL_Platform
	WALL_Window
	WALL_PushWall
	WALL_Switch
	WALL_Door
	SPR_GAD
	SPR_Static
)

const (
	WALLFLAGS_Animated uint32 = 0x1000
	WALLFLAGS_Static   uint32 = 0x2000
	WALLFLAGS_Moving   uint32 = 0x4000
)

const (
	// thin wall directions (e.g. masked walls, entry gates)
	WALLDIR_NorthSouth WallDirection = iota
	WALLDIR_EastWest
)

type ActorInfo struct {
	X                 int
	Y                 int
	WallValue         uint16
	SpriteValue       uint16
	InfoValue         uint16
	Tile              uint16 // matches up to lump name (WALL1, WALL2, etc.)
	Type              ActorType
	MapFlags          uint32
	Damage            bool
	AnimWallID        int // see anim.go
	MaskedWallID      int // see maskedwall.go
	PlatformID        int // see maskedwall.go
	AreaID            int // see area.go
	ThinWallDirection WallDirection
	Item              *ItemInfo
	MapTriggers       []MapTrigger
}

type DoorInfo struct {
	TileID    uint16 // tile sprite ID
	KeyID     uint16 // was there a key sprite for the tiles?
	Direction WallDirection
	Tiles     []*ActorInfo // tiles making up the door
}

type ExitPoint struct {
	X, Y, DestMap int
}

func (actor *ActorInfo) IsWall() bool {
	switch actor.Type {
	case WALL_Regular,
		WALL_ThinWall,
		WALL_Elevator,
		WALL_AnimatedWall,
		WALL_MaskedWall,
		WALL_Platform,
		WALL_Window,
		WALL_PushWall,
		WALL_Door:
		return true
	default:
		return false
	}
}

// html -- true for HTML map gen (return static image equivalent texture name)
//         false for Quake map gen (return Quake animated texture name)
func (actor *ActorInfo) WallTileToTextureName(html bool) string {
	// TODO: correlate with WALLSTRT and EXITSTRT lumps in WAD
	tileId := actor.Tile
	if actor.Type == ACTOR_None {
		return ""
	} else if actor.Type == WALL_Door {
		doorId := 99
		if tileId >= 33 && tileId <= 35 {
			doorId = int(tileId) - 33 + 15
		} else if tileId >= 90 && tileId <= 93 {
			doorId = int(tileId) - 90
		} else if tileId >= 98 && tileId <= 104 {
			doorId = int(tileId) - 90
		} else if tileId >= 94 && tileId <= 97 {
			doorId = int(tileId) - 86
		} else if tileId >= 154 && tileId <= 156 {
			doorId = int(tileId) - 154 + 18
		}
		switch doorId {
		case 0, 8:
			return "RAMDOOR1"
		case 1, 9:
			return "DOOR2"
		case 2, 3, 13:
			return "TRIDOOR1"
		case 10, 11, 14:
			return "SDOOR4"
		case 12:
			return "EDOOR"
		case 15:
			return "SNDOOR"
		case 16:
			return "SNADOOR"
		case 17:
			return "SNKDOOR"
		case 18:
			return "TNDOOR"
		case 19:
			return "TNADOOR"
		case 20:
			return "TNKDOOR"
		default:
			panic(fmt.Sprintf("Illegal door number %d at (%d, %d)", tileId, actor.X, actor.Y))
		}
	} else if actor.Type == WALL_Regular || actor.Type == WALL_ThinWall {
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
			// catch-all, but should never get here
			return fmt.Sprintf("ELEV%d", tileId-71)
		} else if tileId >= 80 && tileId <= 89 {
			return fmt.Sprintf("WALL%d", tileId-16)
		} else {
			return ""
		}
	} else if actor.Type == WALL_AnimatedWall {
		animWallInfo := AnimatedWalls[actor.AnimWallID]
		if html {
			return animWallInfo.StartingLump + "1"
		} else {
			return "+0" + strings.ToLower(animWallInfo.StartingLump)
		}
	} else if html && actor.Type == WALL_MaskedWall {
		maskedWallInfo := MaskedWalls[actor.Tile]
		if maskedWallInfo.IsSwitch {
			return maskedWallInfo.Above
		} else {
			return maskedWallInfo.Bottom
		}
	} else if actor.Type == WALL_Elevator {
		return fmt.Sprintf("ELEV%d", tileId-71)
	} else if html && actor.Type == WALL_Platform {
		return "HSWITCH8"
	} else {
		return ""
	}
}

type RTLMapData struct {
	Header      RTLMapHeader
	WallPlane   [128][128]uint16
	SpritePlane [128][128]uint16
	InfoPlane   [128][128]uint16
	ActorGrid   [128][128]ActorInfo
	Doors       []Door
	ExitPoints  []ExitPoint

	// derived from wall plane
	FloorNumber    int // 0xb4 - 0xc3
	CeilingNumber  int
	Brightness     int
	LightFadeRate  int
	SpawnX         int
	SpawnY         int
	SpawnDirection int

	// derived from sprite plane
	Height     int
	SkyHeight  int
	Fog        int
	IllumWalls int

	// derived from info plane
	SongNumber int

	rtl *RTL
}

func (r *RTLMapData) FloorTexture() string {
	return fmt.Sprintf("FLRCL%d", r.FloorNumber-179)
}

func (r *RTLMapData) CeilingTexture() string {
	if r.CeilingNumber >= 198 && r.CeilingNumber <= 213 {
		return fmt.Sprintf("FLRCL%d", r.CeilingNumber-197)
	}
	return ""
}

func (r *RTLMapData) FloorHeight() int {
	if r.Height >= 90 && r.Height <= 98 {
		return r.Height - 89
	} else if r.Height >= 450 && r.Height <= 457 {
		return r.Height - 441
	} else {
		panic("Map has invalid height")
	}
}

func (r *RTLMapData) CeilingHeight() int {
	if r.Height >= 90 && r.Height <= 98 {
		return r.Height - 89
	} else if r.Height >= 450 && r.Height <= 457 {
		return r.Height - 441
	} else {
		panic("Map has invalid height")
	}
}

// ZOffset -- see rt_stat.c:1091
func (r *RTLMapData) ZOffset(offsetVal uint16, scale float64) float64 {
	if offsetVal&0xff00 != 0xb000 {
		return 0.0
	}
	offsetVal = (offsetVal & 0xff)
	z := offsetVal >> 4
	zf := offsetVal & 0x000f
	if z == 0xf {
		return (float64(zf*4) * scale) - (10.0 * scale)
	} else {
		return ((64.0 + float64(z)*64 + float64(zf)*4) * scale) - (10.0 * scale)
	}
}

// determine which direction thin walls should face
func (r *RTLMapData) ThinWallDirection(x, y int) (WallDirection, int, int) {
	var adjacentCountX, adjacentCountY int

	if x > 0 {
		wallType := r.ActorGrid[y][x-1].Type
		if wallType == WALL_Regular {
			adjacentCountX += 2
		} else if wallType == WALL_MaskedWall {
			adjacentCountX += 4
		} else if wallType != ACTOR_None {
			adjacentCountX++
		}
	}
	if x < 127 {
		wallType := r.ActorGrid[y][x+1].Type
		if wallType == WALL_Regular {
			adjacentCountX += 2
		} else if wallType == WALL_MaskedWall {
			adjacentCountX += 4
		} else if wallType != ACTOR_None {
			adjacentCountX++
		}
	}
	if y > 0 {
		wallType := r.ActorGrid[y-1][x].Type
		if wallType == WALL_Regular {
			adjacentCountY += 2
		} else if wallType == WALL_MaskedWall {
			adjacentCountY += 4
		} else if wallType != ACTOR_None {
			adjacentCountY++
		}
	}
	if y < 127 {
		wallType := r.ActorGrid[y+1][x].Type
		if wallType == WALL_Regular {
			adjacentCountY += 2
		} else if wallType == WALL_MaskedWall {
			adjacentCountY += 4
		} else if wallType != ACTOR_None {
			adjacentCountY++
		}
	}

	if adjacentCountX > adjacentCountY {
		return WALLDIR_EastWest, adjacentCountX, adjacentCountY
	} else {
		return WALLDIR_NorthSouth, adjacentCountX, adjacentCountY
	}
}

type RTL struct {
	fhnd    io.ReadSeeker
	Header  RTLHeader
	MapData [100]RTLMapData
}

func NewRTL(rfile io.ReadSeeker) (*RTL, error) {
	var r RTL
	r.fhnd = rfile

	if err := binary.Read(rfile, binary.LittleEndian, &r.Header); err != nil {
		return nil, err
	}

	if !bytes.Equal(r.Header.Signature[:], rtlMagic[:]) {
		return nil, fmt.Errorf("not an RTL file")
	}

	for i := 0; i < 100; i++ {
		if err := binary.Read(rfile, binary.LittleEndian, &r.MapData[i].Header); err != nil {
			return nil, err
		}

		r.MapData[i].rtl = &r
	}

	for i := 0; i < 100; i++ {
		if err := r.MapData[i].decompressWallPlane(); err != nil {
			return nil, err
		}
		if err := r.MapData[i].decompressSpritePlane(); err != nil {
			return nil, err
		}
		if err := r.MapData[i].decompressInfoPlane(); err != nil {
			return nil, err
		}

		r.MapData[i].FloorNumber = int(r.MapData[i].WallPlane[0][0])
		r.MapData[i].CeilingNumber = int(r.MapData[i].WallPlane[0][1])
		r.MapData[i].Brightness = int(r.MapData[i].WallPlane[0][2])
		r.MapData[i].LightFadeRate = int(r.MapData[i].WallPlane[0][3])

		r.MapData[i].Height = int(r.MapData[i].SpritePlane[0][0])
		r.MapData[i].SkyHeight = int(r.MapData[i].SpritePlane[0][1])
		r.MapData[i].Fog = int(r.MapData[i].SpritePlane[0][2])
		r.MapData[i].IllumWalls = int(r.MapData[i].SpritePlane[0][3])

		r.MapData[i].renderWallGrid()
		r.MapData[i].determineThinWallsAndDirections()
		r.MapData[i].determineMovingWalls()
		r.MapData[i].renderSpriteGrid()
		r.MapData[i].determineExits()
		r.MapData[i].determineGADs()

		for j := 0; j < 128; j++ {
			if r.MapData[i].InfoPlane[0][j]&0xFF00 == 0xBA00 {
				r.MapData[i].SongNumber = int(r.MapData[i].InfoPlane[0][j]) & 0xFF
				break
			}
		}
	}

	return &r, nil
}

func (r *RTLMapData) renderSpriteGrid() {
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			spriteValue := r.SpritePlane[y][x]
			wallValue := r.WallPlane[y][x]

			// spawn location
			if spriteValue >= 19 && spriteValue <= 22 {
				r.SpawnX = x
				r.SpawnY = y
				r.SpawnDirection = int(spriteValue) - 19
			}

			// items (represented as sprites in the RTL data)
			if itemInfo, ok := Items[spriteValue]; ok {
				r.ActorGrid[y][x].Item = &itemInfo
			}

			if wallValue == 0x0b { // fireball shooter
				itemInfo, _ := Items[0x0b]
				r.ActorGrid[y][x].Item = &itemInfo
			}
		}
	}
}

func (r *RTLMapData) renderWallGrid() {
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			// defaults
			r.ActorGrid[y][x].Tile = 0
			r.ActorGrid[y][x].Type = ACTOR_None
			r.ActorGrid[y][x].X = x
			r.ActorGrid[y][x].Y = y

			tileId := r.WallPlane[y][x]
			infoVal := r.InfoPlane[y][x]

			// for reference: rt_ted.c:1965 in ROTT source code (wall
			// setup algorithm, absolute mess)

			// the first 4 tiles are not used and contain metadata.
			// tile values of 0 have nothing
			if (y == 0 && x < 4) || tileId == 0 {
				continue
			}

			if tileId >= AreaTileMin {
				r.ActorGrid[y][x].AreaID = int(tileId - AreaTileMin)
			}

			if (tileId >= 33 && tileId <= 35) || (tileId >= 90 && tileId <= 104) || (tileId >= 154 && tileId <= 156) {
				// doors
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].Type = WALL_Door
				continue
			}

			if tileId <= 32 || (tileId >= 36 && tileId <= 43) {
				// static wall
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Static
				r.ActorGrid[y][x].Type = WALL_Regular
				continue
			} else if tileId >= 72 && tileId <= 79 {
				// elevator tiles
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Static
				r.ActorGrid[y][x].Type = WALL_Elevator
			} else if tileId == 47 || tileId == 48 {
				// TODO: what are these tile ids?
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Static
				r.ActorGrid[y][x].Type = WALL_Regular
			} else if tileId >= 49 && tileId <= 71 {
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Static
				r.ActorGrid[y][x].Type = WALL_Regular
			}

			// for reference: rt_ted.c:2218
			if tileId == 44 || tileId == 45 {
				// animated wall
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Static
				r.ActorGrid[y][x].Type = WALL_AnimatedWall
				if tileId == 44 {
					r.ActorGrid[y][x].Damage = true
					r.ActorGrid[y][x].AnimWallID = 0
				} else {
					r.ActorGrid[y][x].AnimWallID = 3
				}
			} else if tileId == 106 || tileId == 107 {
				// animated wall
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Animated
				r.ActorGrid[y][x].Type = WALL_AnimatedWall
				r.ActorGrid[y][x].AnimWallID = int(tileId) - 105
			} else if tileId >= 224 && tileId <= 233 {
				// animated wall
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Animated
				r.ActorGrid[y][x].Type = WALL_AnimatedWall
				r.ActorGrid[y][x].AnimWallID = int(tileId) - 224 + 4
				if tileId == 233 {
					r.ActorGrid[y][x].Damage = true
				}
			} else if tileId >= 242 && tileId <= 244 {
				// animated wall
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Animated
				r.ActorGrid[y][x].Type = WALL_AnimatedWall
				r.ActorGrid[y][x].AnimWallID = int(tileId) - 242 + 14
			} else if _, ismasked := MaskedWalls[tileId]; ismasked {
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].Type = WALL_MaskedWall
			} else if tileId == 0 || (tileId >= AreaTileMin && tileId <= (AreaTileMin+NumAreas)) {
				// platform
				if infoVal == 1 || (infoVal >= 4 && infoVal <= 9) {
					r.ActorGrid[y][x].Type = WALL_Platform
					r.ActorGrid[y][x].PlatformID = int(infoVal)
					r.ActorGrid[y][x].Tile = tileId
				}
			} else if tileId > 89 || (tileId > 32 && tileId < 36) || tileId == 0 {
				r.ActorGrid[y][x].Tile = 0
				r.ActorGrid[y][x].Type = ACTOR_None
			} else if tileId > 0 && tileId <= 89 {
				r.ActorGrid[y][x].Tile = tileId
				r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Static
				r.ActorGrid[y][x].Type = WALL_Regular
			}

			if r.ActorGrid[y][x].Tile > 1024 {
				log.Fatalf("dun goof at %d, %d (plane: %d)", x, y, tileId)
			}
		}
	}
}

func (r *RTLMapData) determineThinWallsAndDirections() {
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			if r.ActorGrid[y][x].Type == WALL_Regular {
				switch r.ActorGrid[y][x].InfoValue {
				case 1, 4, 5, 6, 7, 8, 9:
					r.ActorGrid[y][x].Type = WALL_ThinWall
					r.ActorGrid[y][x].ThinWallDirection, _, _ = r.ThinWallDirection(x, y)
				}
			}
		}
	}
}

func (r *RTLMapData) determineMovingWalls() {
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			spriteVal := r.ActorGrid[y][x].SpriteValue
			if (r.ActorGrid[y][x].Type == WALL_Regular || r.ActorGrid[y][x].Type == WALL_AnimatedWall) && spriteVal > 0 {
				if _, ok := MoveWallSpriteIDs[spriteVal]; ok {
					// perpetual and/or turbo wall
					r.ActorGrid[y][x].MapFlags |= WALLFLAGS_Moving
					infoVal := r.ActorGrid[y][x].InfoValue
					if infoVal > 0 {
						// touchplate triggered wall
						touchplateX := int((infoVal >> 8) & 0xff)
						touchplateY := int(infoVal & 0xff)
						r.AddTrigger(&r.ActorGrid[y][x], touchplateX, touchplateY, TRIGGER_WallPush)
						//log.Printf("touchplate triggered wall at (%d,%d) has touchplate at (%d,%d)", x, y, touchplateX, touchplateY)
					} else {
						// pushwall
						//log.Printf("pushwall at (%d,%d)", x, y)
					}
				}
			}
		}
	}
}

func (r *RTLMapData) determineExits() {
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			// rt_ted.c:1531
			exitMarker := int(r.ActorGrid[y][x].InfoValue >> 8)
			mapNumber := int(r.ActorGrid[y][x].InfoValue & 0x00ff)
			if exitMarker == 0xe2 || exitMarker == 0xe4 {
				r.ExitPoints = append(r.ExitPoints, ExitPoint{
					X:       x,
					Y:       y,
					DestMap: mapNumber + 1,
				})
			}
		}
	}
}

func (r *RTLMapData) decompressPlane(plane *[128][128]uint16, rlewtag uint32, planeName string) error {
	var curValue uint16

	for i := 0; i < 128*128; {
		if err := binary.Read(r.rtl.fhnd, binary.LittleEndian, &curValue); err != nil {
			return err
		}

		if curValue != uint16(rlewtag) {
			plane[i/128][i%128] = curValue
			actor := &r.ActorGrid[i/128][i%128]
			switch planeName {
			case "wall":
				actor.WallValue = curValue
			case "sprite":
				actor.SpriteValue = curValue
			case "info":
				actor.InfoValue = curValue
			}
			i += 1
		} else {
			var count uint16

			if err := binary.Read(r.rtl.fhnd, binary.LittleEndian, &count); err != nil {
				return err
			}
			if err := binary.Read(r.rtl.fhnd, binary.LittleEndian, &curValue); err != nil {
				return err
			}

			for j := uint16(0); j < count; j++ {
				plane[i/128][i%128] = curValue
				actor := &r.ActorGrid[i/128][i%128]
				switch planeName {
				case "wall":
					actor.WallValue = curValue
				case "sprite":
					actor.SpriteValue = curValue
				case "info":
					actor.InfoValue = curValue
				}
				i += 1
				if i >= 128*128 {
					break
				}
			}
		}
	}

	return nil
}

func (r *RTLMapData) decompressWallPlane() error {
	_, err := r.rtl.fhnd.Seek(int64(r.Header.WallPlaneOffset), io.SeekStart)
	if err != nil {
		return err
	}
	return r.decompressPlane(&r.WallPlane, r.Header.RLEWTag, "wall")
}
func (r *RTLMapData) decompressSpritePlane() error {
	_, err := r.rtl.fhnd.Seek(int64(r.Header.SpritePlaneOffset), io.SeekStart)
	if err != nil {
		return err
	}
	return r.decompressPlane(&r.SpritePlane, r.Header.RLEWTag, "sprite")
}
func (r *RTLMapData) decompressInfoPlane() error {
	_, err := r.rtl.fhnd.Seek(int64(r.Header.InfoPlaneOffset), io.SeekStart)
	if err != nil {
		return err
	}
	return r.decompressPlane(&r.InfoPlane, r.Header.RLEWTag, "info")
}

func (r *RTLMapData) MapName() string {
	return string(bytes.Trim(r.Header.Name[:], "\x00"))
}

func (r *RTLMapData) DumpWallToFile(w io.Writer) error {
	var err error
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			dispValue := r.ActorGrid[y][x]
			if x == r.SpawnX && y == r.SpawnY {
				switch r.SpawnDirection {
				case 0:
					_, err = fmt.Fprintf(w, " P^ ")
				case 1:
					_, err = fmt.Fprintf(w, " P> ")
				case 2:
					_, err = fmt.Fprintf(w, " PV ")
				case 3:
					_, err = fmt.Fprintf(w, " <P ")
				default:
					panic("Bad player direction")
				}
			} else if dispValue.Type != ACTOR_None {
				_, err = fmt.Fprintf(w, " %02x ", dispValue.Tile)
			} else {
				_, err = fmt.Fprintf(w, "    ")
			}
			if err != nil {
				return err
			}
		}
		_, err = w.Write([]byte{'\r', '\n', '\r', '\n'})
		if err != nil {
			return err
		}
	}
	return nil
}

type CellData struct {
	Wall   uint16
	Sprite uint16
	Info   uint16
	Type   string
	X      int
	Y      int
	Img    string
}
type RowData struct {
	Cells []CellData
}

func (r *RTLMapData) DumpMapToHtmlFile(w io.Writer) error {
	tmpl := `
{{define "mapcell"}}<td class="mapcell" id="cell-{{ .X }}-{{ .Y }}" {{ if ne .Img "" }}style="background: url(imgs/{{ .Img }}.png)"{{end}}>
<div>(X:{{ .X }},Y:{{ .Y }}) {{ .Type }}</div>
<div>W: {{ printf "%04x" .Wall }}</div>
<div>S: {{ printf "%04x" .Sprite }}</div>
<div>I: {{ printf "%04x" .Info }}</div>
</td>{{end}}
{{define "maprow"}}<tr class="maprow">{{ range .Cells }}{{ template "mapcell" . }}{{ end }}</tr>{{end}}
{{define "map"}}
<!DOCTYPE html>
<html>
  <head>
    <title>Map</title>
    <style type="text/css">
      .map {
	    border: 0px;
		margin: 0px;
		padding: 0px;
		width: 8452px;
      }
      .maprow {
	  }
	  .maprow > td:hover {
        background: #be5454 !important;
	  }
      .mapcell {
      	width: 64px;
      	height: 64px;
		display: inline-block;
      	border: 1px solid #000;
		padding: 0px;
		margin: 0px
      }
      .mapcell > div {
		font-size: 11px;
		padding: 0px;
		margin: 0px;
      }
    </style>
  </head>
  <body>
    <table class="map">
    {{ range .Rows }}
    {{ template "maprow" . }}
    {{ end }}
    </table>
  </body>
</html>
{{end}}
`

	htmlMapData := struct {
		Rows []RowData
	}{}
	mapTmpl := template.Must(template.New("map").Parse(tmpl))
	for y := 0; y < 128; y++ {
		var cellData []CellData
		for x := 0; x < 128; x++ {
			wallInfo := r.ActorGrid[y][x]
			img := wallInfo.WallTileToTextureName(true)
			switch wallInfo.Type {
			case WALL_Regular, WALL_ThinWall:
				img = "wall/" + img
			case WALL_AnimatedWall:
				img = "anim/" + img
			case WALL_Elevator:
				img = "elev/" + img
			case WALL_Door:
				img = "doors/" + img
			case WALL_MaskedWall, WALL_Platform:
				img = "masked/" + img
			}
			cellData = append(cellData, CellData{
				Wall:   r.WallPlane[y][x],
				Sprite: r.SpritePlane[y][x],
				Info:   r.InfoPlane[y][x],
				Type:   wallInfo.Type.String(),
				X:      x,
				Y:      y,
				Img:    img,
			})
		}
		htmlMapData.Rows = append(htmlMapData.Rows, RowData{Cells: cellData})
	}

	return mapTmpl.ExecuteTemplate(w, "map", htmlMapData)
}

func (r *RTL) PrintMetadata() {
	fmt.Printf("Version: 0x%x\n", r.Header.Version)

	for idx, md := range r.MapData {
		if md.Header.Used == 0 {
			continue
		}
		fmt.Printf("Map #%d\n", idx+1)
		fmt.Printf("\tUsed: %d\n", md.Header.Used)
		fmt.Printf("\tCRC: 0x%x\n", md.Header.CRC)
		fmt.Printf("\tRLEWTag: 0x%x\n", md.Header.RLEWTag)
		fmt.Printf("\tWall Plane Offset: %d\n", md.Header.WallPlaneOffset)
		fmt.Printf("\tSprite Plane Offset: %d\n", md.Header.SpritePlaneOffset)
		fmt.Printf("\tInfo Plane Offset: %d\n", md.Header.InfoPlaneOffset)
		fmt.Printf("\tWall Plane Length: %d\n", md.Header.WallPlaneLength)
		fmt.Printf("\tSprite Plane Length: %d\n", md.Header.SpritePlaneLength)
		fmt.Printf("\tInfo Plane Length: %d\n", md.Header.InfoPlaneLength)
		fmt.Printf("\tMap Name: %s\n", md.MapName())
		fmt.Printf("\tHeight: %d\n", md.FloorHeight())
		fmt.Printf("\tSky Height: %d\n", md.SkyHeight)
	}
}
