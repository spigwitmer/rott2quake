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
	WALL_Regular WallType = iota
	WALL_Elevator
	WALL_AnimatedWall
	WALL_MaskedWall
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
}

type RTLMapData struct {
	Header         RTLMapHeader
	WallPlane      [128][128]uint16
	CookedWallGrid [128][128]WallInfo
	SpritePlane    [128][128]uint16
	InfoPlane      [128][128]uint16

	// derived from wall plane
	FloorNumber   int // 0xb4 - 0xc3
	CeilingNumber int
	Brightness    int
	LightFadeRate int

	// derived from sprite plane
	Height     int
	SkyHeight  int
	Fog        int
	IllumWalls int

	// derived from info plane
	SongNumber int

	rtl *RTL
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
	}

	return &r, nil
}

func (r *RTLMapData) renderWallGrid() {
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			plane := r.WallPlane[i][j]

			if plane <= 32 {
				// static wall
				r.CookedWallGrid[i][j].Tile = plane
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Regular
			} else if plane == 44 || plane == 45 {
				// animated wall
				r.CookedWallGrid[i][j].Tile = plane - 3
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_AnimatedWall
				if plane == 44 {
					r.CookedWallGrid[i][j].Damage = true
					r.CookedWallGrid[i][j].AnimWallID = 0
				} else {
					r.CookedWallGrid[i][j].AnimWallID = 3
				}
			} else if plane == 106 || plane == 107 {
				// animated wall
				r.CookedWallGrid[i][j].Tile = plane - 105
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Animated
				r.CookedWallGrid[i][j].Type = WALL_AnimatedWall
				r.CookedWallGrid[i][j].AnimWallID = int(plane) - 105
			} else if plane >= 224 && plane <= 233 {
				// animated wall
				r.CookedWallGrid[i][j].Tile = plane - 224 + 94
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Animated
				r.CookedWallGrid[i][j].Type = WALL_AnimatedWall
				r.CookedWallGrid[i][j].AnimWallID = int(plane) - 224 + 4
				if plane == 233 {
					r.CookedWallGrid[i][j].Damage = true
				}
			} else if plane >= 242 && plane <= 244 {
				// animated wall
				r.CookedWallGrid[i][j].Tile = plane - 242 + 102
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Animated
				r.CookedWallGrid[i][j].Type = WALL_AnimatedWall
				r.CookedWallGrid[i][j].AnimWallID = int(plane) - 242 + 14
			} else if _, ismasked := MaskedWalls[plane]; ismasked {
				r.CookedWallGrid[i][j].Tile = plane
				r.CookedWallGrid[i][j].Type = WALL_MaskedWall
			} else if plane > 89 || (plane > 32 && plane < 36) {
				r.CookedWallGrid[i][j].Tile = 0
			} else { // (>= 36 && <= 43) || (>= 47 && <= 88)
				// static wall
				r.CookedWallGrid[i][j].Tile = plane - 3
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Regular
			}

			if r.CookedWallGrid[i][j].Tile > 1024 {
				log.Fatalf("dun goof at %d, %d (plane: %d)", i, j, plane)
			}
			if plane > 75 && plane <= 79 {
				r.CookedWallGrid[i][j].Type = WALL_Elevator
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
			if dispValue.Tile > 0 {
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
	}
}
