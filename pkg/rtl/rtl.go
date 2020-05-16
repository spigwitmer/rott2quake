package rtl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

type RTLMapData struct {
	Header      RTLMapHeader
	WallPlane   [128][128]uint16
	SpritePlane [128][128]uint16
	InfoPlane   [128][128]uint16
	rtl         *RTL
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
	}

	return &r, nil
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
			dispValue := r.WallPlane[i][j] & 255
			if dispValue > 0 {
				_, err = fmt.Fprintf(w, " %02x ", dispValue)
			} else {
				_, err = fmt.Fprintf(w, "    ")
			}
			if err != nil {
				return err
			}
		}
		_, err = w.Write([]byte{'\r', '\n'})
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
