package rtl

type TriggerAction int

const (
	TRIGGER_WallPush TriggerAction = iota
)

type MapTrigger struct {
	Actor  *ActorInfo
	Action TriggerAction
}

func (r *RTLMapData) AddTrigger(srcActor *ActorInfo, X, Y int, action TriggerAction) {
	trigger := MapTrigger{
		Actor:  srcActor,
		Action: action,
	}
	r.ActorGrid[Y][X].MapTriggers = append(r.ActorGrid[Y][X].MapTriggers, trigger)
}
