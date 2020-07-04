package rtl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

var (
	rtlMagic = [4]byte{'R', 'T', 'L', '\x00'}
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

type WallType int

const (
	WALL_None WallType = iota
	WALL_Regular
	WALL_Elevator
	WALL_AnimatedWall
	WALL_MaskedWall
	WALL_Window
	WALL_PushWall
)

const (
	WALLFLAGS_Static   uint32 = 0x1000
	WALLFLAGS_Animated uint32 = 0x2000
)

type WallInfo struct {
	Tile         uint16 // matches up to lump name (WALL1, WALL2, etc.)
	Type         WallType
	MapFlags     uint32
	Damage       bool
	AnimWallID   int // see anim.go
	MaskedWallID int // see maskedwall.go
	AreaID       int // see area.go
}

type SpriteInfo struct {
}

type RTLMapData struct {
	Header           RTLMapHeader
	WallPlane        [128][128]uint16
	CookedWallGrid   [128][128]WallInfo
	CookedSpriteGrid [128][128]SpriteInfo
	SpritePlane      [128][128]uint16
	InfoPlane        [128][128]uint16

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

func (r *RTLMapData) FloorHeight() int {
	if r.Height >= 90 && r.Height <= 98 {
		return r.Height - 89
	} else if r.Height >= 450 && r.Height <= 457 {
		return r.Height - 441
	} else {
		panic("Map has invalid height")
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
		r.MapData[i].renderWallGrid()
		r.MapData[i].renderSpriteGrid()

		r.MapData[i].FloorNumber = int(r.MapData[i].WallPlane[0][0])
		r.MapData[i].CeilingNumber = int(r.MapData[i].WallPlane[0][1])
		r.MapData[i].Brightness = int(r.MapData[i].WallPlane[0][2])
		r.MapData[i].LightFadeRate = int(r.MapData[i].WallPlane[0][3])

		r.MapData[i].Height = int(r.MapData[i].SpritePlane[0][0])
		r.MapData[i].SkyHeight = int(r.MapData[i].SpritePlane[0][1])
		r.MapData[i].Fog = int(r.MapData[i].SpritePlane[0][2])
		r.MapData[i].IllumWalls = int(r.MapData[i].SpritePlane[0][3])

		for j := 0; j < 128; j++ {
			if r.MapData[i].InfoPlane[0][j]&0xFF00 == 0xBA00 {
				r.MapData[i].SongNumber = int(r.MapData[i].InfoPlane[0][j]) & 0xFF
				break
			}
		}
		CalculateAreas(&r.MapData[i])
	}

	return &r, nil
}

func (r *RTLMapData) renderSpriteGrid() {
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			spriteValue := r.SpritePlane[i][j]

			// spawn location
			if spriteValue >= 19 && spriteValue <= 22 {
				r.SpawnX = i
				r.SpawnY = j
				r.SpawnDirection = int(spriteValue) - 19
			}
		}
	}
}

func (r *RTLMapData) renderWallGrid() {
	//index := uint16(0)
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			// defaults
			r.CookedWallGrid[i][j].Type = WALL_None

			tileId := r.WallPlane[i][j]

			// for reference: rt_ted.c:1965 in ROTT source code (wall
			// setup algorithm, absolute mess)

			// the first 4 tiles are not used and contain metadata
			if j == 0 && i < 4 {
				r.CookedWallGrid[i][j].Tile = 0
				r.CookedWallGrid[i][j].Type = WALL_None
				continue
			}

			if tileId > 89 || (tileId > 32 && tileId < 36) || tileId == 44 || tileId == 45 || tileId == 0 {
				r.CookedWallGrid[i][j].Tile = 0
				r.CookedWallGrid[i][j].Type = WALL_None
				continue
			}

			/*
				if tileId <= 32 {
					index = tileId
				} else {
					index = tileId - 3
				}
			*/

			if tileId <= 32 {
				// static wall
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Regular
			} else if tileId > 75 && tileId <= 79 {
				// elevator tiles
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Elevator
			} else if tileId == 47 || tileId == 48 {
				// TODO: what are these tile ids?
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Regular
			} else {
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Regular
			}

			// TODO: animated wall masking, heights
			if tileId == 44 || tileId == 45 {
				// animated wall
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_AnimatedWall
				if tileId == 44 {
					r.CookedWallGrid[i][j].Damage = true
					r.CookedWallGrid[i][j].AnimWallID = 0
				} else {
					r.CookedWallGrid[i][j].AnimWallID = 3
				}
			} else if tileId == 106 || tileId == 107 {
				// animated wall
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Animated
				r.CookedWallGrid[i][j].Type = WALL_AnimatedWall
				r.CookedWallGrid[i][j].AnimWallID = int(tileId) - 105
			} else if tileId >= 224 && tileId <= 233 {
				// animated wall
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Animated
				r.CookedWallGrid[i][j].Type = WALL_AnimatedWall
				r.CookedWallGrid[i][j].AnimWallID = int(tileId) - 224 + 4
				if tileId == 233 {
					r.CookedWallGrid[i][j].Damage = true
				}
			} else if tileId >= 242 && tileId <= 244 {
				// animated wall
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Animated
				r.CookedWallGrid[i][j].Type = WALL_AnimatedWall
				r.CookedWallGrid[i][j].AnimWallID = int(tileId) - 242 + 14
			} else if _, ismasked := MaskedWalls[tileId]; ismasked {
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].Type = WALL_MaskedWall
			} else if (tileId >= 36 && tileId <= 43) || (tileId >= 47 && tileId <= 88) {
				// static wall
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Regular
			}

			if r.CookedWallGrid[i][j].Tile > 1024 {
				log.Fatalf("dun goof at %d, %d (plane: %d)", i, j, tileId)
			}
		}
	}
}

func (r *RTLMapData) decompressPlane(plane *[128][128]uint16, rlewtag uint32) error {
	var curValue uint16

	for i := 0; i < 128*128; {
		if err := binary.Read(r.rtl.fhnd, binary.LittleEndian, &curValue); err != nil {
			return err
		}

		if curValue != uint16(rlewtag) {
			plane[i/128][i%128] = curValue
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
	return r.decompressPlane(&r.WallPlane, r.Header.RLEWTag)
}
func (r *RTLMapData) decompressSpritePlane() error {
	_, err := r.rtl.fhnd.Seek(int64(r.Header.SpritePlaneOffset), io.SeekStart)
	if err != nil {
		return err
	}
	return r.decompressPlane(&r.SpritePlane, r.Header.RLEWTag)
}
func (r *RTLMapData) decompressInfoPlane() error {
	_, err := r.rtl.fhnd.Seek(int64(r.Header.InfoPlaneOffset), io.SeekStart)
	if err != nil {
		return err
	}
	return r.decompressPlane(&r.InfoPlane, r.Header.RLEWTag)
}

func (r *RTLMapData) MapName() string {
	return string(bytes.Trim(r.Header.Name[:], "\x00"))
}

func (r *RTLMapData) DumpWallToFile(w io.Writer) error {
	var err error
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			dispValue := r.CookedWallGrid[i][j]
			if i == r.SpawnX && j == r.SpawnY {
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
					panic("How did this happen?")
				}
			} else if dispValue.Type != WALL_None {
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
		fmt.Printf("\tHeight: %d\n", md.Height)
		fmt.Printf("\tSky Height: %d\n", md.SkyHeight)
	}
}
