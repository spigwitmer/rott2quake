package rtl

type AnimatedWallInfo struct {
	Ticks        int
	NumFrames    int
	StartingLump string
}

// ripped straight out of rt_stat.c from ROTT
var AnimatedWalls = []AnimatedWallInfo{
	{3, 4, "FPLACE1"},   //lava wall
	{3, 6, "ANIMY1"},    //anim red
	{3, 6, "ANIMR1"},    //anim yellow
	{40, 4, "ANIMFAC1"}, //anim face
	{3, 4, "ANIMONE1"},  //anim one
	{3, 4, "ANIMTWO1"},  //anim two
	{3, 4, "ANIMTHR1"},  //anim three
	{3, 4, "ANIMFOR1"},  //anim four
	{3, 6, "ANIMGW1"},   //anim grey water
	{3, 6, "ANIMYOU1"},  //anim you do not belong
	{3, 6, "ANIMBW1"},   //anim brown water
	{3, 6, "ANIMBP1"},   //anim brown piston
	{3, 6, "ANIMCHN1"},  //anim chain
	{3, 6, "ANIMFW1"},   //anim firewall
	{3, 6, "ANIMLAT1"},  //anim little blips
	{3, 6, "ANIMST1"},   //anim light streams left
	{3, 6, "ANIMRP1"},   //anim light streams right
}
