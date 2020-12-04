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

var (
	DIR_East      int = 0
	DIR_Northeast int = 1
	DIR_North     int = 2
	DIR_Northwest int = 3
	DIR_West      int = 4
	DIR_Southwest int = 5
	DIR_South     int = 6
	DIR_Southeast int = 7
	DIR_Unknown   int = 8
	ICONARROWS        = 72 // rt_actor.c:11010

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
	InitialDirection int
}

type PathNode struct {
	X         int
	Y         int
	Direction int
	Next      *PathNode
}

func (r *RTLMapData) DetermineWallPath(actor *ActorInfo) (WallPathType, *PathNode) {
	var nodes []*PathNode
	markedNodes := make(map[string]*PathNode)

	if actor.Type != WALL_Regular {
		return PATH_Unknown, nil
	}
	addNode := func(X int, Y int, direction int) {
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
		addNode(curX, curY, moveWallInfo.InitialDirection)
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
				panic("I'M FREE!!")
			}
			markerTag := fmt.Sprintf("%d-%d", curX, curY)
			if prevNode, ok := markedNodes[markerTag]; ok {
				nodes[len(nodes)-1].Next = prevNode
				pathType = PATH_Perpetual
				continue
			}
			spriteVal := r.ActorGrid[curX][curY].SpriteValue
			if spriteVal >= 72 && spriteVal <= 78 {
				switch int(spriteVal - 72) {
				case DIR_East, DIR_North, DIR_West, DIR_South:
					curDirection = int(spriteVal - 72)
					addNode(curX, curY, curDirection)
				default:
					panic(fmt.Sprintf("weird direction: %d", spriteVal-72))
				}
			}
		}
	}
	return pathType, nodes[0]
}
