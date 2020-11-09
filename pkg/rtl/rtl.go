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

type WallType int

const (
	WALL_None WallType = iota
	WALL_Regular
	WALL_Elevator
	WALL_AnimatedWall
	WALL_MaskedWall
	WALL_Platform
	WALL_Window
	WALL_PushWall
)

const (
	WALLFLAGS_Animated uint32 = 0x1000
	WALLFLAGS_Static   uint32 = 0x2000
)

type WallDirection int

const (
	// thin wall directions (e.g. masked walls, entry gates)
	WALLDIR_NorthSouth WallDirection = iota
	WALLDIR_EastWest
)

type WallInfo struct {
	Tile         uint16 // matches up to lump name (WALL1, WALL2, etc.)
	Type         WallType
	MapFlags     uint32
	Damage       bool
	AnimWallID   int // see anim.go
	MaskedWallID int // see maskedwall.go
	PlatformID   int // see maskedwall.go
	AreaID       int // see area.go
}

// html -- true for HTML map gen (return static image equivalent texture name)
//         false for Quake map gen (return Quake animated texture name)
func (wallInfo *WallInfo) WallTileToTextureName(html bool) string {
	// TODO: correlate with WALLSTRT and EXITSTRT lumps in WAD
	tileId := wallInfo.Tile
	if wallInfo.Type == WALL_None {
		return ""
	} else if wallInfo.Type == WALL_Regular {
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
	} else if wallInfo.Type == WALL_AnimatedWall {
		animWallInfo := AnimatedWalls[wallInfo.AnimWallID]
		if html {
			return animWallInfo.StartingLump + "1"
		} else {
			return "+0" + strings.ToLower(animWallInfo.StartingLump)
		}
	} else if wallInfo.Type == WALL_Elevator {
		return fmt.Sprintf("ELEV%d", tileId-71)
	} else {
		return ""
	}
}

type SpriteInfo struct {
	Item *ItemInfo
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

// determine which direction thin walls should face
func (r *RTLMapData) ThinWallDirection(x, y int) WallDirection {
	var adjacentCountX, adjacentCountY int

	if x > 0 {
		if r.CookedWallGrid[x-1][y].Type == WALL_Regular {
			adjacentCountX += 2
        } else if r.CookedWallGrid[x-1][y].Type == WALL_MaskedWall {
			adjacentCountX += 2
		} else if r.CookedWallGrid[x-1][y].Type != WALL_None {
			adjacentCountX++
		}
	}
	if x < 127 {
		if r.CookedWallGrid[x+1][y].Type == WALL_Regular {
			adjacentCountX += 2
        } else if r.CookedWallGrid[x+1][y].Type == WALL_MaskedWall {
			adjacentCountX += 2
		} else if r.CookedWallGrid[x+1][y].Type != WALL_None {
			adjacentCountX++
		}
	}
	if y > 0 {
		if r.CookedWallGrid[x][y-1].Type == WALL_Regular {
			adjacentCountY += 2
        } else if r.CookedWallGrid[x][y-1].Type == WALL_MaskedWall {
			adjacentCountY += 2
		} else if r.CookedWallGrid[x][y-1].Type != WALL_None {
			adjacentCountY++
		}
	}
	if y < 127 {
		if r.CookedWallGrid[x][y+1].Type == WALL_Regular {
			adjacentCountY += 2
        } else if r.CookedWallGrid[x][y+1].Type == WALL_MaskedWall {
			adjacentCountY += 2
		} else if r.CookedWallGrid[x][y+1].Type != WALL_None {
			adjacentCountY++
		}
	}

	if adjacentCountX > adjacentCountY {
		return WALLDIR_EastWest
	} else {
		return WALLDIR_NorthSouth
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
	}

	return &r, nil
}

func (r *RTLMapData) renderSpriteGrid() {
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			spriteValue := r.SpritePlane[i][j]
			wallValue := r.WallPlane[i][j]

			// spawn location
			if spriteValue >= 19 && spriteValue <= 22 {
				r.SpawnX = i
				r.SpawnY = j
				r.SpawnDirection = int(spriteValue) - 19
			}

			// items (represented as sprites in the RTL data)
			if itemInfo, ok := Items[spriteValue]; ok {
				r.CookedSpriteGrid[i][j].Item = &itemInfo
			}

			if wallValue == 0x0b { // fireball shooter
				itemInfo, _ := Items[0x0b]
				r.CookedSpriteGrid[i][j].Item = &itemInfo
			}
		}
	}
}

func (r *RTLMapData) renderWallGrid() {
	for i := 0; i < 128; i++ {
		for j := 0; j < 128; j++ {
			// defaults
			r.CookedWallGrid[i][j].Type = WALL_None

			tileId := r.WallPlane[i][j]
			infoVal := r.InfoPlane[i][j]

			// for reference: rt_ted.c:1965 in ROTT source code (wall
			// setup algorithm, absolute mess)

			// the first 4 tiles are not used and contain metadata.
			// tile values of 0 have nothing
			if (i == 0 && j < 4) || tileId == 0 {
				r.CookedWallGrid[i][j].Tile = 0
				r.CookedWallGrid[i][j].Type = WALL_None
				continue
			}

			if tileId >= AreaTileMin {
				r.CookedWallGrid[i][j].AreaID = int(tileId - AreaTileMin)
			}

			if tileId <= 32 || (tileId >= 36 && tileId <= 43) {
				// static wall
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Regular
				continue
			} else if tileId >= 72 && tileId <= 79 {
				// elevator tiles
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Elevator
			} else if tileId == 47 || tileId == 48 {
				// TODO: what are these tile ids?
				r.CookedWallGrid[i][j].Tile = tileId
				r.CookedWallGrid[i][j].MapFlags |= WALLFLAGS_Static
				r.CookedWallGrid[i][j].Type = WALL_Regular
			}

			// for reference: rt_ted.c:2218
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
			} else if (tileId == 0 || (tileId >= AreaTileMin && tileId <= (AreaTileMin+NumAreas))) && infoVal > 0 {
				// platform
				r.CookedWallGrid[i][j].Type = WALL_Platform
				r.CookedWallGrid[i][j].PlatformID = int(infoVal)
				r.CookedWallGrid[i][j].Tile = tileId
			} else if tileId > 89 || (tileId > 32 && tileId < 36) || tileId == 0 {
				r.CookedWallGrid[i][j].Tile = 0
				r.CookedWallGrid[i][j].Type = WALL_None
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
					panic("Bad player direction")
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

type CellData struct {
	Wall   uint16
	Sprite uint16
	Info   uint16
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
<div>({{ .X }}, {{ .Y }})</div>
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
	for i := 0; i < 128; i++ {
		var cellData []CellData
		for j := 0; j < 128; j++ {
			wallInfo := r.CookedWallGrid[i][j]
			img := wallInfo.WallTileToTextureName(true)
			if wallInfo.Type == WALL_Regular {
				img = "wall/" + img
			} else if wallInfo.Type == WALL_AnimatedWall {
				img = "anim/" + img
			} else if wallInfo.Type == WALL_Elevator {
				img = "elev/" + img
			}
			cellData = append(cellData, CellData{
				Wall:   r.WallPlane[i][j],
				Sprite: r.SpritePlane[i][j],
				Info:   r.InfoPlane[i][j],
				X:      i,
				Y:      j,
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
