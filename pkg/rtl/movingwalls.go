package rtl

import (
	"fmt"
	"log"
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
	ICONARROWS    int           = 72 // rt_actor.c:11010

	// rt_ted.c:2984
	MoveWallSpriteIDs = map[uint16]MoveWallInfo{
		// pushwalls
		uint16(DIR_East) + uint16(ICONARROWS):      MoveWallInfo{2.0, DIR_East},
		uint16(DIR_North) + uint16(ICONARROWS):     MoveWallInfo{2.0, DIR_North},
		uint16(DIR_West) + uint16(ICONARROWS):      MoveWallInfo{2.0, DIR_West},
		uint16(DIR_South) + uint16(ICONARROWS):     MoveWallInfo{2.0, DIR_South},
		uint16(DIR_Northeast) + uint16(ICONARROWS): MoveWallInfo{2.0, DIR_Northeast},
		uint16(DIR_Northwest) + uint16(ICONARROWS): MoveWallInfo{2.0, DIR_Northwest},
		uint16(DIR_Southwest) + uint16(ICONARROWS): MoveWallInfo{2.0, DIR_Southwest},
		uint16(DIR_Southeast) + uint16(ICONARROWS): MoveWallInfo{2.0, DIR_Southeast},

		// movewalls
		300: MoveWallInfo{2.0, DIR_East},
		318: MoveWallInfo{2.0, DIR_North},
		336: MoveWallInfo{2.0, DIR_West},
		354: MoveWallInfo{2.0, DIR_South},

		// turbo movewalls
		256: MoveWallInfo{4.0, DIR_East},
		257: MoveWallInfo{4.0, DIR_North},
		258: MoveWallInfo{4.0, DIR_West},
		259: MoveWallInfo{4.0, DIR_South},

		// GADs
		MovingGADEast:  MoveWallInfo{2.0, DIR_East},
		MovingGADNorth: MoveWallInfo{2.0, DIR_North},
		MovingGADWest:  MoveWallInfo{2.0, DIR_West},
		MovingGADSouth: MoveWallInfo{2.0, DIR_South},
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

func (r *RTLMapData) DetermineWallPath(actor *ActorInfo, pushWall bool) (WallPathType, *PathNode, int) {
	var nodes []*PathNode
	markedNodes := make(map[string]*PathNode)

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
	deltaX, deltaY := 0, 0
	pathType := PATH_Unknown
	delta := 1
	if pushWall {
		delta = 2
	}
	if moveWallInfo, ok := MoveWallSpriteIDs[actor.SpriteValue]; ok {
		curDirection := moveWallInfo.InitialDirection
		for pathType == PATH_Unknown {
			switch curDirection {
			case DIR_East:
				deltaX = delta
				deltaY = 0
			case DIR_Northeast:
				deltaX = delta
				deltaY = -delta
			case DIR_North:
				deltaX = 0
				deltaY = -delta
			case DIR_Northwest:
				deltaX = -delta
				deltaY = -delta
			case DIR_West:
				deltaX = -delta
				deltaY = 0
			case DIR_Southwest:
				deltaX = -delta
				deltaY = delta
			case DIR_South:
				deltaX = 0
				deltaY = delta
			case DIR_Southeast:
				deltaX = delta
				deltaY = delta
			default:
				panic("Unknown direction")
			}
			curX += deltaX
			curY += deltaY
			if curX > 127 || curX < 0 || curY > 127 || curY < 0 {
				// I'M FREE!!
				pathType = PATH_Terminal
				addNode(curX, curY, DIR_Unknown)
				switch actor.Type {
				case WALL_Regular:
					log.Printf("wall starting at (%d,%d) IS FREE!! (%d,%d)", actor.X, actor.Y, curX, curY)
				case SPR_GAD:
					log.Printf("GAD starting at (%d,%d) IS FREE!! (%d,%d)", actor.X, actor.Y, curX, curY)
				default:
					log.Printf("object starting at (%d,%d) IS FREE!! (%d,%d)", actor.X, actor.Y, curX, curY)
				}
				continue
			}

			// ran into previous pivot point?
			markerTag := fmt.Sprintf("%d-%d", curX, curY)
			if prevNode, ok := markedNodes[markerTag]; ok {
				nodes[len(nodes)-1].Next = prevNode
				pathType = PATH_Perpetual
				continue
			}

			spriteVal := r.ActorGrid[curY][curX].SpriteValue
			if actor.IsWall() && (spriteVal >= 72 && spriteVal <= 79) {
				switch WallDirection(spriteVal - 72) {
				case DIR_East, DIR_North, DIR_West, DIR_South,
					DIR_Northeast, DIR_Northwest, DIR_Southwest, DIR_Southeast:
					curDirection = WallDirection(spriteVal - 72)
					addNode(curX, curY, curDirection)
				default:
					panic(fmt.Sprintf("(%d,%d) weird direction: %d", curX, curY, spriteVal-72))
				}
			} else if actor.Type == SPR_GAD {
				switch WallDirection(spriteVal - 72) {
				case DIR_East, DIR_North, DIR_West, DIR_South,
					DIR_Northeast, DIR_Northwest, DIR_Southwest, DIR_Southeast:
					curDirection = WallDirection(spriteVal - 72)
					addNode(curX, curY, curDirection)
				default:
					continue
				}
			} else if pushWall {
				addNode(curX, curY, DIR_Unknown)
				pathType = PATH_Terminal
				continue
			}
		}
	}
	if len(nodes) > 0 {
		return pathType, nodes[0], len(nodes)
	} else {
		return pathType, nil, 0
	}
}
