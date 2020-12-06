package quakemap

// create a cube(-ish) brush with 6 sides, no rotation,
// and a single texture
// x/y/z coordinate sets are 2 opposite corners
func BasicCuboid(x1, y1, z1, x2, y2, z2 float64, texture string, scale float64, wrapTexture bool) Brush {
	var b Brush
	var wrapFactor float64 = 1.0

	if wrapTexture {
		wrapFactor = -1.0
	}

	if x1 > x2 {
		tmp := x1
		x1 = x2
		x2 = tmp
	}
	if y1 > y2 {
		tmp := y1
		y1 = y2
		y2 = tmp
	}
	if z1 > z2 {
		tmp := z1
		z1 = z2
		z2 = tmp
	}

	// south
	b.AddPlane(
		x1, y1, z1, // p1
		x1, y1, z1+1, // p2
		x1+1, y1, z1, // p3
		texture,
		0, 0, // offset
		0, // rotation
		scale, scale)
	// north
	b.AddPlane(
		x1, y2, z1, // p1
		x1+1, y2, z1, // p2
		x1, y2, z1+1, // p3
		texture,
		0, 0, // offset
		0, // rotation
		scale*wrapFactor, scale)
	// west
	b.AddPlane(
		x1, y1, z1, // p1
		x1, y1+1, z1, // p2
		x1, y1, z1+1, // p3
		texture,
		0, 0, // offset
		0, // rotation
		scale*wrapFactor, scale)
	// east
	b.AddPlane(
		x2, y1, z1, // p1
		x2, y1, z1+1, // p2
		x2, y1+1, z1, // p3
		texture,
		0, 0, // offset
		0, // rotation
		scale, scale)
	// top
	b.AddPlane(
		x1, y1, z2, // p1
		x1, y1+1, z2, // p2
		x1+1, y1, z2, // p3
		texture,
		0, 0, // offset
		0, // rotation
		scale, scale)
	// bottom
	b.AddPlane(
		x1, y1, z1, // p1
		x1+1, y1, z1, // p2
		x1, y1+1, z1, // p3
		texture,
		0, 0, // offset
		0, // rotation
		scale, scale)

	return b
}
