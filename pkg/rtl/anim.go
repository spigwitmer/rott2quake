package rtl

import (
	"fmt"
)

type AnimatedWallInfo struct {
	Ticks        int
	NumFrames    int
	StartingLump string
}

// ripped straight out of rt_stat.c from ROTT
var AnimatedWalls = []AnimatedWallInfo{
	{3, 4, "FPLACE"},   //lava wall
	{3, 6, "ANIMY"},    //anim red
	{3, 6, "ANIMR"},    //anim yellow
	{40, 4, "ANIMFAC"}, //anim face
	{3, 4, "ANIMONE"},  //anim one
	{3, 4, "ANIMTWO"},  //anim two
	{3, 4, "ANIMTHR"},  //anim three
	{3, 4, "ANIMFOR"},  //anim four
	{3, 6, "ANIMGW"},   //anim grey water
	{3, 6, "ANIMYOU"},  //anim you do not belong
	{3, 6, "ANIMBW"},   //anim brown water
	{3, 6, "ANIMBP"},   //anim brown piston
	{3, 6, "ANIMBP"},   //anim brown piston
	//{3, 6, "ANIMCHN"},  //anim chain (not actually used)
	{3, 6, "ANIMFW"},  //anim firewall
	{3, 6, "ANIMLAT"}, //anim little blips
	{3, 6, "ANIMST"},  //anim light streams left
	{3, 6, "ANIMRP"},  //anim light streams right
}

// returns whether the lump is an animated wall frame + the frame number
func GetAnimatedWallInfo(lumpName string) (*AnimatedWallInfo, int) {
	for _, walldata := range AnimatedWalls {
		for i := 1; i <= walldata.NumFrames; i++ {
			if fmt.Sprintf("%s%d", walldata.StartingLump, i) == lumpName {
				return &walldata, i
			}
		}
	}
	return nil, 0
}
