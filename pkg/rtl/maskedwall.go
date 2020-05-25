package rtl

// TODO: platforms

type MaskedWallInfo struct {
	Flags uint16
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

var MaskedWalls = map[uint16]MaskedWallInfo{
	MW_HiSwitchOff:         MaskedWallInfo{MWF_Blocking},
	MW_MultiGlass1:         MaskedWallInfo{MWF_Multi | MWF_Blocking | MWF_BlockingChanges | MWF_Shootable},
	MW_MultiGlass2:         MaskedWallInfo{MWF_Multi | MWF_Blocking | MWF_BlockingChanges | MWF_Shootable},
	MW_MultiGlass3:         MaskedWallInfo{MWF_Multi | MWF_Blocking | MWF_BlockingChanges | MWF_Shootable},
	MW_ShotOutGlass1:       MaskedWallInfo{MWF_BottomPassable},
	MW_ShotOutGlass2:       MaskedWallInfo{MWF_BottomPassable},
	MW_ShotOutGlass3:       MaskedWallInfo{MWF_BottomPassable},
	MW_Normal1:             MaskedWallInfo{MWF_Blocking},
	MW_Normal1Shootable:    MaskedWallInfo{MWF_Blocking | MWF_Shootable},
	MW_Normal2:             MaskedWallInfo{MWF_Blocking},
	MW_Normal2Shootable:    MaskedWallInfo{MWF_Blocking | MWF_Shootable},
	MW_Normal3:             MaskedWallInfo{MWF_Blocking},
	MW_Normal3Shootable:    MaskedWallInfo{MWF_Blocking | MWF_Shootable},
	MW_SinglePaneShootable: MaskedWallInfo{MWF_Shootable | MWF_BlockingChanges | MWF_Blocking},
	MW_SinglePane:          MaskedWallInfo{MWF_BottomPassable},
	MW_DogWall:             MaskedWallInfo{MWF_NonDogBlocking | MWF_WeaponBlocking},
	MW_PeepHole:            MaskedWallInfo{MWF_WeaponBlocking | MWF_Blocking},
	MW_ExitArch:            MaskedWallInfo{MWF_BottomPassable},
	MW_SecretExitArch:      MaskedWallInfo{MWF_BottomPassable},
	MW_EntryGate:           MaskedWallInfo{MWF_Blocking},
	MW_HiSwitchOn:          MaskedWallInfo{MWF_Blocking | MWF_SwitchOn},
	MW_Railing:             MaskedWallInfo{MWF_AbovePassable | MWF_MiddlePassable},
}
