package rtl

import (
	"fmt"
)

type WallPathType int

var (
	PATH_Unknown   WallPathType = 0
	PATH_Perpetual WallPathType = 1
	PATH_Terminal  WallPathType = 2
)

type WallDirection int

func (w *WallDirection) Name() string {
	switch int(*w) {
	case 0:
		return "East"
	case 1:
		return "Northeast"
	case 2:
		return "North"
	case 3:
		return "Northwest"
	case 4:
		return "West"
	case 5:
		return "Southwest"
	case 6:
		return "South"
	case 7:
		return "Southeast"
	default:
		return "???"
	}
}

var (
	DIR_East      WallDirection = 0
	DIR_Northeast WallDirection = 1
	DIR_North     WallDirection = 2
	DIR_Northwest WallDirection = 3
	DIR_West      WallDirection = 4
	DIR_Southwest WallDirection = 5
	DIR_South     WallDirection = 6
	DIR_Southeast WallDirection = 7
	DIR_Unknown   WallDirection = 8
	ICONARROWS                  = 72 // rt_actor.c:11010

	// rt_ted.c:2984
	MoveWallSpriteIDs = map[uint16]MoveWallInfo{
		300: MoveWallInfo{2.0, DIR_East},
		318: MoveWallInfo{2.0, DIR_North},
		336: MoveWallInfo{2.0, DIR_West},
		354: MoveWallInfo{2.0, DIR_South},
		256: MoveWallInfo{4.0, DIR_East},
		257: MoveWallInfo{4.0, DIR_North},
		258: MoveWallInfo{4.0, DIR_West},
		259: MoveWallInfo{4.0, DIR_South},
	}
)

type MoveWallInfo struct {
	Speed            int
	InitialDirection WallDirection
}

type PathNode struct {
	X         int
	Y         int
	Direction WallDirection
	Next      *PathNode
}

func (r *RTLMapData) DetermineWallPath(actor *ActorInfo) (WallPathType, *PathNode, int) {
	var nodes []*PathNode
	markedNodes := make(map[string]*PathNode)

	if actor.Type != WALL_Regular {
		return PATH_Unknown, nil, 0
	}
	addNode := func(X int, Y int, direction WallDirection) {
		p := PathNode{X: X, Y: Y, Direction: direction, Next: nil}
		markerTag := fmt.Sprintf("%d-%d", X, Y)
		markedNodes[markerTag] = &p
		if len(nodes) > 0 {
			nodes[len(nodes)-1].Next = &p
		}
		nodes = append(nodes, &p)
	}
	curX, curY := actor.X, actor.Y
	pathType := PATH_Unknown
	if moveWallInfo, ok := MoveWallSpriteIDs[actor.SpriteValue]; ok {
		curDirection := moveWallInfo.InitialDirection
		for pathType == PATH_Unknown {
			switch curDirection {
			case DIR_East:
				curX++
			case DIR_North:
				curY--
			case DIR_West:
				curX--
			case DIR_South:
				curY++
			default:
				panic("Unknown direction")
			}
			if curX > 127 || curX < 0 || curY > 127 || curY < 0 {
				// I'M FREE!!
				pathType = PATH_Terminal
			}
			markerTag := fmt.Sprintf("%d-%d", curX, curY)
			if prevNode, ok := markedNodes[markerTag]; ok {
				nodes[len(nodes)-1].Next = prevNode
				pathType = PATH_Perpetual
				continue
			}
			spriteVal := r.ActorGrid[curY][curX].SpriteValue
			if spriteVal >= 72 && spriteVal <= 78 {
				switch WallDirection(spriteVal - 72) {
				case DIR_East, DIR_North, DIR_West, DIR_South:
					curDirection = WallDirection(spriteVal - 72)
					addNode(curX, curY, curDirection)
				default:
					panic(fmt.Sprintf("weird direction: %d", spriteVal-72))
				}
			}
		}
	}
	if len(nodes) > 0 {
		return pathType, nodes[0], len(nodes)
	} else {
		return pathType, nil, 0
	}
}
