package quakemap

type FaceParams struct {
	Texture              string
	TexScaleX, TexScaleY float64
}

// a cube(-ish) brush with 6 sides, no rotation,
// and a single texture
type CuboidParams struct {
	North, South, East, West, Top, Bottom FaceParams
}

func BasicCuboidParams(texture string, scale float64, wrapTexture bool) CuboidParams {
	var params CuboidParams

	applyParams := func(destFace *FaceParams, faceTexture string, faceScale float64) {
		destFace.Texture = faceTexture
		destFace.TexScaleX = faceScale
		destFace.TexScaleY = faceScale
	}

	applyParams(&params.North, texture, scale)
	applyParams(&params.South, texture, scale)
	applyParams(&params.East, texture, scale)
	applyParams(&params.West, texture, scale)
	applyParams(&params.Top, texture, scale)
	applyParams(&params.Bottom, texture, scale)

	if wrapTexture {
		applyParams(&params.North, texture, -scale)
		applyParams(&params.West, texture, -scale)
	}

	return params
}

// x/y/z coordinate sets are 2 opposite corners
func BasicCuboid(x1, y1, z1, x2, y2, z2 float64, texture string, scale float64, wrapTexture bool) Brush {
	return BuildCuboidBrush(x1, y1, z1, x2, y2, z2, BasicCuboidParams(texture, scale, wrapTexture))
}

func BuildCuboidBrush(x1, y1, z1, x2, y2, z2 float64, params CuboidParams) Brush {
	var b Brush

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
		x1, y1, z2, // p1
		x1+1, y1, z2, // p2
		x1, y1, z2-1, // p3
		params.South.Texture,
		0, 0, // offset
		0, // rotation
		params.South.TexScaleX, params.South.TexScaleY)
	// north
	b.AddPlane(
		x2, y2, z2, // p1
		x2-1, y2, z2, // p2
		x2, y2, z2-1, // p3
		params.North.Texture,
		0, 0, // offset
		0, // rotation
		params.North.TexScaleX, params.North.TexScaleY)
	// west
	b.AddPlane(
		x1, y2, z2, // p1
		x1, y2-1, z2, // p2
		x1, y2, z2-1, // p3
		params.West.Texture,
		0, 0, // offset
		0, // rotation
		params.West.TexScaleX, params.West.TexScaleY)
	// east
	b.AddPlane(
		x2, y1, z2, // p1
		x2, y1+1, z2, // p2
		x2, y1, z2-1, // p3
		params.East.Texture,
		0, 0, // offset
		0, // rotation
		params.East.TexScaleX, params.East.TexScaleY)
	// top
	b.AddPlane(
		x1, y1, z2, // p1
		x1, y1+1, z2, // p2
		x1+1, y1, z2, // p3
		params.Top.Texture,
		0, 0, // offset
		0, // rotation
		params.Top.TexScaleX, params.Top.TexScaleY)
	// bottom
	b.AddPlane(
		x1, y1, z1, // p1
		x1+1, y1, z1, // p2
		x1, y1+1, z1, // p3
		params.Bottom.Texture,
		0, 0, // offset
		0, // rotation
		params.Bottom.TexScaleX, params.Bottom.TexScaleY)

	return b
}
