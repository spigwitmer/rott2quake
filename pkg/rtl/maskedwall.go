package rtl

// TODO: platforms

type MaskedWallInfo struct {
	// wall properties
	Flags uint16
	// lump names for each wall component
	Side, Middle, Above, Bottom string
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
	MW_HiSwitchOff:         MaskedWallInfo{MWF_Blocking, "", "HSWITCH2", "HSWITCH3", "HSWITCH1"},
	MW_MultiGlass1:         MaskedWallInfo{MWF_Multi | MWF_Blocking | MWF_BlockingChanges | MWF_Shootable, "SIDE21", "ABOVEM5A", "ABOVEM5", "MULTI1A"},
	MW_MultiGlass2:         MaskedWallInfo{MWF_Multi | MWF_Blocking | MWF_BlockingChanges | MWF_Shootable, "SIDE21", "ABOVEM5B", "ABOVEM5", "MULTI2A"},
	MW_MultiGlass3:         MaskedWallInfo{MWF_Multi | MWF_Blocking | MWF_BlockingChanges | MWF_Shootable, "SIDE21", "ABOVEM5C", "ABOVEM5", "MULTI3A"},
	MW_ShotOutGlass1:       MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM5A", "ABOVEM5", "MULTI1"},
	MW_ShotOutGlass2:       MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM5B", "ABOVEM5", "MULTI2"},
	MW_ShotOutGlass3:       MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM5C", "ABOVEM5", "MULTI3"},
	MW_Normal1:             MaskedWallInfo{MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED1"},
	MW_Normal1Shootable:    MaskedWallInfo{MWF_Blocking | MWF_Shootable, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED1A"},
	MW_Normal2:             MaskedWallInfo{MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED2"},
	MW_Normal2Shootable:    MaskedWallInfo{MWF_Blocking | MWF_Shootable, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED2A"},
	MW_Normal3:             MaskedWallInfo{MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED3"},
	MW_Normal3Shootable:    MaskedWallInfo{MWF_Blocking | MWF_Shootable, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED3A"},
	MW_SinglePaneShootable: MaskedWallInfo{MWF_Shootable | MWF_BlockingChanges | MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED4A"},
	MW_SinglePane:          MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM4A", "ABOVEM4", "MASKED4"},
	MW_DogWall:             MaskedWallInfo{MWF_NonDogBlocking | MWF_WeaponBlocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "DOGMASK"},
	MW_PeepHole:            MaskedWallInfo{MWF_WeaponBlocking | MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "PEEPMASK"},
	MW_ExitArch:            MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM4A", "ABOVEM4", "EXITARCH"},
	MW_SecretExitArch:      MaskedWallInfo{MWF_BottomPassable, "SIDE21", "ABOVEM4A", "ABOVEM4", "EXITARCA"},
	MW_EntryGate:           MaskedWallInfo{MWF_Blocking, "SIDE21", "ABOVEM4A", "ABOVEM4", "ENTRARCH"},
	MW_HiSwitchOn:          MaskedWallInfo{MWF_Blocking | MWF_SwitchOn, "", "HSWITCH2", "HSWITCH4", "HSWITCH1"},
	MW_Railing:             MaskedWallInfo{MWF_AbovePassable | MWF_MiddlePassable, "", "", "", "RAILING"},
}
