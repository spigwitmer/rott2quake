package rtl

type TouchplateAction int

const (
	TOUCH_WallPush TouchplateAction = iota
)

type TouchplateTrigger struct {
	Actor  *ActorInfo
	Action TouchplateAction
}

func (r *RTLMapData) AddTouchplateTrigger(srcActor *ActorInfo, X, Y int, action TouchplateAction) {
	trigger := TouchplateTrigger{
		Actor:  srcActor,
		Action: action,
	}
	r.ActorGrid[Y][X].TouchplateTriggers = append(r.ActorGrid[Y][X].TouchplateTriggers, trigger)
}
