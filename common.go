package olion

type Coordinates struct {
	X int
	Y int
	Z int
}

//func (c *Coordinates) Add(c1, c2 *Coordinates) *Coordinates {
//	c = &Coordinates{X: c1.X + c2.X, Y: c1.Y + C2.Y, Z: c1.Z + C2.Z}
//	return c
//}

func (c1 Coordinates) Add(c2 Coordinates) Coordinates {
	return Coordinates{X: c1.X + c2.X, Y: c1.Y + c2.Y, Z: c1.Z + c2.Z}
}

func (c Coordinates) ScaleBy(k int) Coordinates {
	return Coordinates{X: c.X * k, Y: c.Y * k, Z: c.Z * k}
}

func (c Coordinates) Div(k int) Coordinates {
	return Coordinates{X: c.X / k, Y: c.Y / k, Z: c.Z / k}
}

/*
type Direction struct {
	theta float64
	phi   float64
}
*/
