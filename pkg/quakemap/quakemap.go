package quakemap

import (
	"fmt"
	"math"
	"strings"
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

func (p *Plane) Clone() Plane {
	var notp Plane
	notp.X1 = p.X1
	notp.X2 = p.X2
	notp.X3 = p.X3
	notp.Y1 = p.Y1
	notp.Y2 = p.Y2
	notp.Y3 = p.Y3
	notp.Z1 = p.Z1
	notp.Z2 = p.Z2
	notp.Z3 = p.Z3
	notp.Texture = p.Texture
	notp.Xoffset = p.Xoffset
	notp.Yoffset = p.Yoffset
	notp.Rotation = p.Rotation
	notp.Xscale = p.Xscale
	notp.Yscale = p.Yscale
	return notp
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

func (b *Brush) Clone() Brush {
	var notb Brush
	for _, plane := range b.Planes {
		notb.Planes = append(notb.Planes, plane.Clone())
	}
	return notb
}

// determine how much of the X axis the brush spans across
func (b *Brush) Width() float64 {
	var vertMax, vertMin float64

	vertMax = math.Inf(-1)
	vertMin = math.Inf(1)

	setMinOrMax := func(val float64) {
		if val > vertMax {
			vertMax = val
		} else if val < vertMin {
			vertMin = val
		}
	}

	for _, plane := range b.Planes {
		setMinOrMax(plane.X1)
		setMinOrMax(plane.X2)
		setMinOrMax(plane.X3)
	}

	return vertMax - vertMin
}

// determine how much of the Y axis the brush takes
func (b *Brush) Length() float64 {
	var vertMax, vertMin float64

	vertMax = math.Inf(-1)
	vertMin = math.Inf(1)

	setMinOrMax := func(val float64) {
		if val > vertMax {
			vertMax = val
		} else if val < vertMin {
			vertMin = val
		}
	}

	for _, plane := range b.Planes {
		setMinOrMax(plane.Y1)
		setMinOrMax(plane.Y2)
		setMinOrMax(plane.Y3)
	}

	return vertMax - vertMin
}

// determine how much of the Z axis the brush takes
func (b *Brush) Height() float64 {
	var vertMax, vertMin float64

	vertMax = math.Inf(-1)
	vertMin = math.Inf(1)

	setMinOrMax := func(val float64) {
		if val > vertMax {
			vertMax = val
		} else if val < vertMin {
			vertMin = val
		}
	}

	for _, plane := range b.Planes {
		setMinOrMax(plane.Z1)
		setMinOrMax(plane.Z2)
		setMinOrMax(plane.Z3)
	}

	return vertMax - vertMin
}

// scale the plane vertices with (cx, cy, cz) as the focal point
func (b *Brush) Scale(cx, cy, cz, scale float64) {
	doScale := func(sx, sy, sz *float64) {
		*sx = (*sx - cx) * scale
		*sy = (*sy - cy) * scale
		*sz = (*sz - cz) * scale
	}

	for i, _ := range b.Planes {
		doScale(&b.Planes[i].X1, &b.Planes[i].Y1, &b.Planes[i].Z1)
		doScale(&b.Planes[i].X2, &b.Planes[i].Y2, &b.Planes[i].Z2)
		doScale(&b.Planes[i].X3, &b.Planes[i].Y3, &b.Planes[i].Z3)
	}
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

func (b *Brush) Translate(x, y, z float64) {
	for i, _ := range b.Planes {
		b.Planes[i].X1 += x
		b.Planes[i].X2 += x
		b.Planes[i].X3 += x
		b.Planes[i].Y1 += y
		b.Planes[i].Y2 += y
		b.Planes[i].Y3 += y
		b.Planes[i].Z1 += z
		b.Planes[i].Z2 += z
		b.Planes[i].Z3 += z
	}
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

func (e *Entity) Height() float64 {
	var vertMax, vertMin float64

	vertMax = math.Inf(-1)
	vertMin = math.Inf(1)

	setMinOrMax := func(val float64) {
		if val > vertMax {
			vertMax = val
		} else if val < vertMin {
			vertMin = val
		}
	}

	for _, brush := range e.Brushes {
		for _, plane := range brush.Planes {
			setMinOrMax(plane.Z1)
			setMinOrMax(plane.Z2)
			setMinOrMax(plane.Z3)
		}
	}

	return vertMax - vertMin
}

func (e *Entity) Render() string {
	output := fmt.Sprintf(`{
"spawnflags" "%d"
"classname" "%s"
`, e.SpawnFlags, e.ClassName)

	for k, v := range e.AdditionalKeys {
		output += fmt.Sprintf("\"%s\" \"%s\"\n", k, v)
	}

	if e.ClassName != "worldspawn" {
		output += fmt.Sprintf("\"origin\" \"%.02f %.02f %.02f\"\n",
			e.OriginX, e.OriginY, e.OriginZ)
		if e.ClassName == "info_player_start" || e.Angle != 0 {
			output += fmt.Sprintf("\"angle\" \"%.02f\"\n", e.Angle)
		}
	}

	switch e.ClassName {
	case "worldspawn":
		output += fmt.Sprintf("\"wad\" \"%s\"\n", strings.Join(e.Map.Wads, ";"))
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
	Wads            []string
	WorldSpawn      *Entity
	InfoPlayerStart *Entity
	Entities        []*Entity
}

func NewQuakeMap(startx, starty, startz float64) *QuakeMap {
	var qmap QuakeMap

	qmap.WorldSpawn = NewEntity(0, "worldspawn", &qmap)
	qmap.WorldSpawn.AdditionalKeys["light"] = "256"
	qmap.InfoPlayerStart = NewEntity(0, "info_player_start", &qmap)
	qmap.InfoPlayerStart.OriginX = startx
	qmap.InfoPlayerStart.OriginY = starty
	qmap.InfoPlayerStart.OriginZ = startz

	return &qmap
}

func (q *QuakeMap) Render() string {
	output := q.WorldSpawn.Render() + "\n" + q.InfoPlayerStart.Render()
	for _, entity := range q.Entities {
		output += "\n" + entity.Render()
	}
	return output
}
