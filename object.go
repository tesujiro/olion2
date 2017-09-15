package olion

type Part_type int

type Part struct {
	dots []Coordinates
	// color
}

type Parter interface {
	getDots() []Coordinates
	addDot(Coordinates)
}

/*
func newPart() Part {
	return Part{}
}
*/

func (p Part) getDots() []Coordinates {
	return p.dots
}

func (p Part) addDot(d Coordinates) {
	p.dots = append(p.dots, d)
}

/* one dot part */
type DotPart struct {
	Part
}

func newDotPart(p Parter) DotPart {
	// Todo: check len(p.getDots())
	return DotPart{
		Part: p.(Part),
	}
}

type LinePart struct {
	Part
}

type RectanglePart struct {
	Part
	fill bool
}

func newRectanglePart(p Parter) RectanglePart {
	// Todo: check len(p.getDots())
	return RectanglePart{
		Part: p.(Part),
	}
}

type Object struct {
	parts    []Parter
	position Coordinates //位置
	//Direction Direction   //方向
	//Speed
	//created
	//weight

	size int
}

type Shaper interface {
	shape() []Parter
	addPart(Parter)
}

func (obj *Object) shape() []Parter {
	return obj.parts
}

func (obj *Object) addPart(p Parter) {
	obj.parts = append(obj.parts, p)
}

type Star struct {
	Object
}

func newStar(s int, c Coordinates) *Star {
	star := Star{}
	star.size = s
	star.position = c
	dot := newDotPart(Part{
		//Type: Part_Dot,
		dots: []Coordinates{
			Coordinates{X: 0, Y: 0, Z: 0},
		},
	})
	star.addPart(dot)
	return &star
}

type SpaceShip struct {
	Object
}

func newSpaceShip(s int, c Coordinates) *SpaceShip {
	ship := SpaceShip{}
	ship.size = s
	ship.position = c
	rectangle := newRectanglePart(Part{
		//Type: Part_Rectangle,
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
		},
	})
	ship.addPart(rectangle)
	return &ship
}
