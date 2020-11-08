package quakemap

// create a cube(-ish) brush with 6 sides, no rotation,
// and a single texture
// x/y/z coordinate sets are 2 opposite corners
func BasicCuboid(x1, y1, z1, x2, y2, z2 float64, texture string, scale float64) Brush {
	var b Brush

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
		scale*-1.0, scale)
	// west
	b.AddPlane(
		x1, y1, z1, // p1
		x1, y1+1, z1, // p2
		x1, y1, z1+1, // p3
		texture,
		0, 0, // offset
		0, // rotation
		scale*-1.0, scale)
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
