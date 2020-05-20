package quakemap

import (
	"fmt"
	//"regexp"
	//"strings"
)

type Brush struct {
	X1, Y1, Z1       float64
	X2, Y2, Z2       float64
	X3, Y3, Z3       float64
	Texture          string
	Xoffset, Yoffset float64
	Rotation         float64
	Xscale, Yscale   float64
}

func (b Brush) Render() string {
	texture := b.Texture
	if texture == "" {
		texture = "__TB_empty"
	}
	return fmt.Sprintf("(%.02f %.02f %.02f) (%.02f %.02f %.02f) (%.02f %.02f %.02f) %s %.02f %.02f %.02f %.02f %.02f",
		b.X1, b.Y1, b.Z1,
		b.X2, b.Y2, b.Z2,
		b.X3, b.Y3, b.Z3,
		texture,
		b.Xoffset, b.Yoffset,
		b.Rotation,
		b.Xscale, b.Yscale,
	)
}

type Entity struct {
	SpawnFlags int
	ClassName  string
	Brushes    []Brush

	// for info_player_start
	OriginX float64
	OriginY float64
	OriginZ float64
}

func (e Entity) Render() string {
	output := fmt.Sprintf(`{
    "spawnflags" "%d"
    "classname" "%s"`, e.SpawnFlags, e.ClassName)

	if e.ClassName == "info_player_start" {
		output += fmt.Sprintf("\n    \"origin\" \"%.02f %.02f %.02f\"",
			e.OriginX, e.OriginY, e.OriginZ)
	}

	if len(e.Brushes) > 0 {
		output += "\n    {"
		for _, brush := range e.Brushes {
			output += "\n        " + brush.Render()
		}
		output += "\n    }"
	}

	output += "\n}"
	return output
}

type QuakeMap struct {
	Wad             string
	WorldSpawn      Entity
	InfoPlayerStart Entity
	Entities        []Entity
}

func NewQuakeMap(startx, starty, startz float64) *QuakeMap {
	var qmap QuakeMap

	qmap.WorldSpawn.SpawnFlags = 0
	qmap.WorldSpawn.ClassName = "worldspawn"
	qmap.InfoPlayerStart.SpawnFlags = 0
	qmap.InfoPlayerStart.ClassName = "info_player_start"
	qmap.InfoPlayerStart.OriginX = startx
	qmap.InfoPlayerStart.OriginY = starty
	qmap.InfoPlayerStart.OriginZ = startz

	return &qmap
}

/*
func indent(what string, byhowmuch int) string {
	re := regexp.MustCompile(`^`)
	return re.ReplaceAllLiteralString(what, strings.Repeat(" ", byhowmuch))
}
*/

func (q QuakeMap) Render() string {
	output := q.WorldSpawn.Render() + "\n" + q.InfoPlayerStart.Render()
	for _, entity := range q.Entities {
		output += "\n" + entity.Render()
	}
	return output
}
