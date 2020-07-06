package quakemap

import (
	"fmt"
)

type Plane struct {
	X1, Y1, Z1       float64
	X2, Y2, Z2       float64
	X3, Y3, Z3       float64
	Texture          string
	Xoffset, Yoffset float64
	Rotation         float64
	Xscale, Yscale   float64
}

func (p *Plane) Render() string {
	texture := p.Texture
	if texture == "" {
		texture = "__TB_empty"
	}
	return fmt.Sprintf("(%.02f %.02f %.02f) (%.02f %.02f %.02f) (%.02f %.02f %.02f) %s %.02f %.02f %.02f %.02f %.02f",
		p.X1, p.Y1, p.Z1,
		p.X2, p.Y2, p.Z2,
		p.X3, p.Y3, p.Z3,
		texture,
		p.Xoffset, p.Yoffset,
		p.Rotation,
		p.Xscale, p.Yscale,
	)
}

type Brush struct {
	Planes []Plane
}

func (b *Brush) AddPlane(
	x1, y1, z1, x2, y2, z2, x3, y3, z3 float64,
	texture string,
	xOffset, yOffset float64,
	rotation float64,
	xScale, yScale float64) {

	plane := Plane{
		X1:       x1,
		Y1:       y1,
		Z1:       z1,
		X2:       x2,
		Y2:       y2,
		Z2:       z2,
		X3:       x3,
		Y3:       y3,
		Z3:       z3,
		Texture:  texture,
		Xoffset:  xOffset,
		Yoffset:  yOffset,
		Rotation: rotation,
		Xscale:   xScale,
		Yscale:   yScale,
	}
	b.Planes = append(b.Planes, plane)
}

func (b *Brush) Render() string {
	out := "{\n"
	for _, plane := range b.Planes {
		out += plane.Render() + "\n"
	}
	out += "}\n"
	return out
}

type Entity struct {
	SpawnFlags     int
	ClassName      string
	Brushes        []Brush
	Map            *QuakeMap
	AdditionalKeys map[string]string

	// for info_player_start
	OriginX float64
	OriginY float64
	OriginZ float64
	Angle   float64
}

func NewEntity(spawnFlags int, className string, qm *QuakeMap) *Entity {
	var e Entity
	e.SpawnFlags = spawnFlags
	e.ClassName = className
	e.AdditionalKeys = make(map[string]string)
	e.Map = qm
	return &e
}

func (e *Entity) Render() string {
	output := fmt.Sprintf(`{
"spawnflags" "%d"
"classname" "%s"
`, e.SpawnFlags, e.ClassName)

	for k, v := range e.AdditionalKeys {
		output += fmt.Sprintf("\"%s\" \"%s\"\n", k, v)
	}

	switch e.ClassName {
	case "info_player_start":
		output += fmt.Sprintf("\"origin\" \"%.02f %.02f %.02f\"\n",
			e.OriginX, e.OriginY, e.OriginZ)
		output += fmt.Sprintf("\"angle\" \"%.02f\"\n", e.Angle)
	case "worldspawn":
		output += fmt.Sprintf("\"wad\" \"%s\"\n", e.Map.Wad)
	}

	if len(e.Brushes) > 0 {
		for idx, brush := range e.Brushes {
			output += fmt.Sprintf("// brush %d\n", idx)
			output += brush.Render() + "\n"
		}
	}
	output += "}\n"
	return output
}

type QuakeMap struct {
	Wad             string
	WorldSpawn      *Entity
	InfoPlayerStart *Entity
	Entities        []*Entity
}

func NewQuakeMap(startx, starty, startz float64) *QuakeMap {
	var qmap QuakeMap

	qmap.WorldSpawn = NewEntity(0, "worldspawn", &qmap)
	qmap.InfoPlayerStart = NewEntity(0, "info_player_start", &qmap)
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

func (q *QuakeMap) Render() string {
	output := q.WorldSpawn.Render() + "\n" + q.InfoPlayerStart.Render()
	for _, entity := range q.Entities {
		output += "\n" + entity.Render()
	}
	return output
}
