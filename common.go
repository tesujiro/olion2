package olion

type Coordinates struct {
	X int
	Y int
	Z int
}

/*
func between(a, b, c int) bool {
	return (a < b && b <= c) || (a > b && b >= c)
}

func (c1 Coordinates) between(c0, c2 Coordinates) bool {
	return between(c0.X, c1.X, c2.X) && between(c0.Y, c1.Y, c2.Y) && between(c0.Z, c1.Z, c2.Z)
}
*/

func (c1 Coordinates) Add(c2 Coordinates) Coordinates {
	return Coordinates{X: c1.X + c2.X, Y: c1.Y + c2.Y, Z: c1.Z + c2.Z}
}

func (c Coordinates) ScaleBy(k int) Coordinates {
	return Coordinates{X: c.X * k, Y: c.Y * k, Z: c.Z * k}
}

func (c Coordinates) Div(k int) Coordinates {
	return Coordinates{X: c.X / k, Y: c.Y / k, Z: c.Z / k}
}

func (center Coordinates) Symmetry(diff Coordinates) (ret []Coordinates) {
	/*
		perm := func(a, b int) (ret []int) {
			if b == 0 {
				ret = []int{a}
			} else {
				ret = []int{a - b, a + b}
			}
			return ret
		}
		for _, x := range perm(center.X, diff.X) {
			for _, y := range perm(center.Y, diff.Y) {
				for _, z := range perm(center.Z, diff.Z) {
					ret = append(ret, Coordinates{X: x, Y: y, Z: z})
				}
			}
		}
	*/
	switch {
	case diff.X == 0:
		ret = []Coordinates{
			center.Add(Coordinates{X: 0, Y: diff.Y, Z: diff.Z}),
			center.Add(Coordinates{X: 0, Y: diff.Y, Z: -diff.Z}),
			center.Add(Coordinates{X: 0, Y: -diff.Y, Z: -diff.Z}),
			center.Add(Coordinates{X: 0, Y: -diff.Y, Z: diff.Z}),
		}
	case diff.Y == 0:
		ret = []Coordinates{
			center.Add(Coordinates{X: diff.X, Y: 0, Z: diff.Z}),
			center.Add(Coordinates{X: diff.X, Y: 0, Z: -diff.Z}),
			center.Add(Coordinates{X: -diff.X, Y: 0, Z: -diff.Z}),
			center.Add(Coordinates{X: -diff.X, Y: 0, Z: diff.Z}),
		}
	case diff.Z == 0:
		ret = []Coordinates{
			center.Add(Coordinates{X: diff.X, Y: diff.Y, Z: 0}),
			center.Add(Coordinates{X: diff.X, Y: -diff.Y, Z: 0}),
			center.Add(Coordinates{X: -diff.X, Y: -diff.Y, Z: 0}),
			center.Add(Coordinates{X: -diff.X, Y: diff.Y, Z: 0}),
		}
	default:
	}
	//debug.Printf("Symmetry=%v\n", ret)
	return ret
}

/*
type Direction struct {
	theta float64
	phi   float64
}
*/
