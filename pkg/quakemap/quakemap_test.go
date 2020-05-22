package quakemap

import (
	"testing"
)

func TestQuakeMap(t *testing.T) {
	q := NewQuakeMap(7, 8, 9)
	t.Log(q.Render())
}

func TestQuakeMapWithBrushes(t *testing.T) {
	q := NewQuakeMap(7, 8, 9)
	b := Brush{
		X1: 3.0, Y1: 4.0, Z1: 5.0,
		X2: 13.0, Y2: 14.0, Z2: 15.0,
		X3: 23.0, Y3: 24.0, Z3: 25.0,
		Texture:  "footexbar",
		Xoffset:  33.0,
		Yoffset:  34.0,
		Rotation: 43.0,
		Xscale:   53.0,
		Yscale:   64.0,
	}
	e := Entity{
		SpawnFlags: 0,
		ClassName:  "foobar",
		Brushes:    []Brush{b},
	}
	q.Entities = append(q.Entities, e)
	t.Log(q.Render())
}
