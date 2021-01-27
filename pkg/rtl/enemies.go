package rtl

type Difficulty int

var (
	DifficultyEasy Difficulty = 1
	DifficultyHard Difficulty = 2
)

type EnemyConversionInfo struct {
	QuakeEnemyNames []string
	DuskEnemyNames  []string
}

func (e *EnemyConversionInfo) EntityName(actor *ActorInfo, dusk bool) string {
	// provide some variety in the enemies spawned but keep it
	// deterministic
	enemyPool := e.QuakeEnemyNames
	if dusk {
		enemyPool = e.DuskEnemyNames
	}
	if len(enemyPool) == 0 {
		return ""
	}
	return enemyPool[(actor.X+actor.Y)%len(enemyPool)]
}

type EnemyInfo struct {
	Direction      int // 0-359
	Difficulty     Difficulty
	ConversionInfo EnemyConversionInfo
}

var Enemies = map[string]EnemyConversionInfo{
	"low_guard": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_army"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"sneaky_low_guard": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_ogre"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"high_guard": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_ogre_marksman"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"overpatrol_guard": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_wizard"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"triad_enforcer": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_shambler"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"lightning_guard": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_demon1"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"monk": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_knight"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"fire_monk": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_hellknight"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"robo_guard": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_enforcer"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"ballistikraft": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_enforcer"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"gun_emplacement": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_dog"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	"4_way_gun": EnemyConversionInfo{
		QuakeEnemyNames: []string{"monster_dog"},
		DuskEnemyNames:  []string{"monster_leatherneck"},
	},
	// TODO: bosses
}

func GetEnemyInfoFromSpriteValue(spriteValue uint16) *EnemyInfo {
	var enemyName string
	var enemyInfo EnemyInfo
	var difficulty Difficulty

	// rt_ted.c:4592
	switch s := spriteValue; {
	case s >= 108 && s <= 119:
		enemyName = "low_guard"
		difficulty = DifficultyEasy
	case s >= 126 && s <= 137:
		enemyName = "low_guard"
		difficulty = DifficultyHard
	case s == 120:
		enemyName = "sneaky_low_guard"
		difficulty = DifficultyEasy
	case s == 138:
		enemyName = "sneaky_low_guard"
		difficulty = DifficultyHard
	case s >= 144 && s <= 155:
		enemyName = "high_guard"
		difficulty = DifficultyEasy
	case s >= 162 && s <= 173:
		enemyName = "high_guard"
		difficulty = DifficultyHard
	case s >= 216 && s <= 227:
		enemyName = "overpatrol_guard"
		difficulty = DifficultyEasy
	case s >= 234 && s <= 245:
		enemyName = "overpatrol_guard"
		difficulty = DifficultyHard
	case s >= 180 && s <= 191:
		enemyName = "strike_guard"
		difficulty = DifficultyEasy
	case s >= 198 && s <= 204:
		enemyName = "strike_guard"
		difficulty = DifficultyHard
	case s >= 288 && s <= 299:
		enemyName = "triad_enforcer"
		difficulty = DifficultyEasy
	case s >= 306 && s <= 317:
		enemyName = "triad_enforcer"
		difficulty = DifficultyHard
	case s >= 324 && s <= 335:
		enemyName = "lightning_guard"
		difficulty = DifficultyEasy
	case s >= 342 && s <= 353:
		enemyName = "lightning_guard"
		difficulty = DifficultyHard
	case s >= 360 && s <= 371:
		enemyName = "monk"
		difficulty = DifficultyEasy
	case s >= 378 && s <= 389:
		enemyName = "monk"
		difficulty = DifficultyHard
	case s >= 396 && s <= 407:
		enemyName = "fire_monk"
		difficulty = DifficultyEasy
	case s >= 414 && s <= 425:
		enemyName = "fire_monk"
		difficulty = DifficultyHard
	case s >= 158 && s <= 161:
		enemyName = "robo_guard"
		difficulty = DifficultyEasy
	case s >= 176 && s <= 179:
		enemyName = "robo_guard"
		difficulty = DifficultyHard
	case s >= 408 && s <= 411:
		enemyName = "ballistikraft"
		difficulty = DifficultyEasy
	case s >= 426 && s <= 429:
		enemyName = "ballistikraft"
		difficulty = DifficultyHard
	case s >= 194 && s <= 197:
		enemyName = "gun_emplacement"
		difficulty = DifficultyEasy
	case s >= 212 && s <= 215:
		enemyName = "gun_emplacement"
		difficulty = DifficultyHard
	case s == 89:
		enemyName = "4_way_gun"
		difficulty = DifficultyEasy
	case s == 211:
		enemyName = "4_way_gun"
		difficulty = DifficultyHard
	default:
		return nil
	}

	enemyInfo.Difficulty = difficulty
	enemyInfo.ConversionInfo = Enemies[enemyName]
	// TODO: direction
	return &enemyInfo
}
