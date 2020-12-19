package rtl

// TODO: platforms

type MaskedWallInfo struct {
	// wall properties
	Flags uint16
	// lump names for each wall component
	Side, Middle, Above, Bottom string
	// is it a switch?
	IsSwitch bool
}

const (
	MWF_Shootable = uint16(1) << iota
	MWF_Blocking
	MWF_Multi
	MWF_BlockingChanges
	MWF_AbovePassable
	MWF_NonDogBlocking
	MWF_WeaponBlocking
	MWF_BottomPassable
	MWF_MiddlePassable
	MWF_ABP
	MWF_SwitchOn
	MWF_BottomFlipping
	MWF_TopFlipping
)

const (
	MW_HiSwitchOff         = uint16(157)
	MW_MultiGlass1         = uint16(158)
	MW_MultiGlass2         = uint16(159)
	MW_MultiGlass3         = uint16(160)
	MW_Normal1Shootable    = uint16(162)
	MW_Normal1             = uint16(163)
	MW_Normal2Shootable    = uint16(164)
	MW_Normal2             = uint16(165)
	MW_Normal3Shootable    = uint16(166)
	MW_Normal3             = uint16(167)
	MW_SinglePaneShootable = uint16(168)
	MW_SinglePane          = uint16(169)
	MW_DogWall             = uint16(170)
	MW_PeepHole            = uint16(171)
	MW_ExitArch            = uint16(172)
	MW_SecretExitArch      = uint16(173)
	MW_EntryGate           = uint16(174)
	MW_HiSwitchOn          = uint16(175)
	MW_ShotOutGlass1       = uint16(176)
	MW_ShotOutGlass2       = uint16(177)
	MW_ShotOutGlass3       = uint16(178)
	MW_Railing             = uint16(179)
)

// rt_door.c:2211
// TODO: this only accounts for the registered version, shareware
// differs from this
var MaskedWalls = map[uint16]MaskedWallInfo{
	MW_HiSwitchOff:         MaskedWallInfo{MWF_Blocking, "", "HSWITCH2", "HSWITCH3", "HSWITCH1", true},
	MW_MultiGlass1:         MaskedWallInfo{MWF_Multi | MWF_Blocking | MWF_BlockingChanges | MWF_Shootable, "SIDE21", "ABOVEM5A", "ABOVEM5", "MULTI1A", false},
	MW_MultiGlass2:         MaskedWallInfo{MWF_Multi | MWF_Blocking | MWF_BlockingChanges | MWF_Shootable, "SIDE21", "ABOVEM5B", "ABOVEM5", "MULTI2A", false},
	MW_MultiGlass3:         MaskedWallInfo{MWF_Multi | MWF_Blocking | MWF_BlockingChanges | MWF_Shootable, "SIDE21", "ABOVEM5C", "ABOVEM5", "MULTI3A", false},
	MW_ShotOutGlass1:       MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM5A", "ABOVEM5", "MULTI1", false},
	MW_ShotOutGlass2:       MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM5B", "ABOVEM5", "MULTI2", false},
	MW_ShotOutGlass3:       MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM5C", "ABOVEM5", "MULTI3", false},
	MW_Normal1:             MaskedWallInfo{MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED1", false},
	MW_Normal1Shootable:    MaskedWallInfo{MWF_Blocking | MWF_Shootable, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED1A", false},
	MW_Normal2:             MaskedWallInfo{MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED2", false},
	MW_Normal2Shootable:    MaskedWallInfo{MWF_Blocking | MWF_Shootable, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED2A", false},
	MW_Normal3:             MaskedWallInfo{MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED3", false},
	MW_Normal3Shootable:    MaskedWallInfo{MWF_Blocking | MWF_Shootable, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED3A", false},
	MW_SinglePaneShootable: MaskedWallInfo{MWF_Shootable | MWF_BlockingChanges | MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED4A", false},
	MW_SinglePane:          MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED4", false},
	MW_DogWall:             MaskedWallInfo{MWF_NonDogBlocking | MWF_WeaponBlocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "DOGMASK", false},
	MW_PeepHole:            MaskedWallInfo{MWF_WeaponBlocking | MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "PEEPMASK", false},
	MW_ExitArch:            MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM4A", "ABOVEM4", "EXITARCH", false},
	MW_SecretExitArch:      MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM4A", "ABOVEM4", "EXITARCA", false},
	MW_EntryGate:           MaskedWallInfo{MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "ENTRARCH", false},
	// the typo is intentional, the same typo is in DARKWAR.WAD
	MW_HiSwitchOn: MaskedWallInfo{MWF_Blocking | MWF_SwitchOn, "", "HSWITCH2", "HSWTICH4", "HSWITCH1", true},
	MW_Railing:    MaskedWallInfo{MWF_AbovePassable | MWF_MiddlePassable, "", "", "", "RAILING", false},
}

var HMSK_Lumps = []string{
	"HSWITCH1",
	"HSWITCH2",
	"HSWITCH3",
	"HSWTICH4",
	"HSWITCH5",
	"HSWITCH6",
	"HSWTCH6A",
	"HSWITCH7",
	"HSWITCH8",
	"HSWTCH8A",
	"HSWTCH9",
	"HSWTCH10",
	"HSWTCH11",
	"HSWTCH12",
	"HSWTCH13",
	"HSWTCH14",
}

// rt_ted.c:2818
var Platforms = map[int]MaskedWallInfo{
	4: MaskedWallInfo{MWF_BottomPassable | MWF_MiddlePassable, "", "", HMSK_Lumps[10], "", false},
	5: MaskedWallInfo{MWF_AbovePassable | MWF_MiddlePassable, "", "", "", HMSK_Lumps[8], false},
	6: MaskedWallInfo{MWF_MiddlePassable, "", "", HMSK_Lumps[10], HMSK_Lumps[8], false},
	7: MaskedWallInfo{MWF_BottomPassable, "", HMSK_Lumps[7], HMSK_Lumps[7], HMSK_Lumps[12], false},
	8: MaskedWallInfo{MWF_BottomPassable | MWF_AbovePassable, "", HMSK_Lumps[7], HMSK_Lumps[5], HMSK_Lumps[12], false},
	9: MaskedWallInfo{MWF_AbovePassable, "", HMSK_Lumps[7], HMSK_Lumps[5], HMSK_Lumps[4], false},
	1: MaskedWallInfo{MWF_AbovePassable, "", HMSK_Lumps[7], HMSK_Lumps[5], HMSK_Lumps[4], false},
}
