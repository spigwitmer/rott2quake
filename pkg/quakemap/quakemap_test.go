package quakemap

import (
	"testing"
)

func TestQuakeMap(t *testing.T) {
	q := NewQuakeMap(7, 8, 9)
	t.Log(q.Render())
}
