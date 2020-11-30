package rtl

// ZOffset -- see rt_stat.c:1091
func ZOffset(offsetVal uint16) int {
	offsetVal = (offsetVal & 0xff)
	z := offsetVal >> 4
	zf := offsetVal & 0xf
	if zf == 0xf {
		return 64 - int(zf<<2)
	} else {
		return 0 - int(z<<6) - int(zf<<2)
	}
}
