package rtl

// "Gravitational Anomaly Disks"

const (
	// rt_ted.c:3744
	StaticGAD      uint16 = 0x01cd
	ElevatingGAD   uint16 = 0x01ce
	MovingGADEast  uint16 = 0x01cf
	MovingGADNorth uint16 = 0x01d0
	MovingGADWest  uint16 = 0x01d1
	MovingGADSouth uint16 = 0x01d2
)

func (r *RTLMapData) determineGADs() {
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			actor := r.ActorGrid[y][x]
			if actor.SpriteValue >= StaticGAD && actor.SpriteValue <= MovingGADSouth {
				r.ActorGrid[y][x].Type = SPR_GAD
			}
		}
	}
}
