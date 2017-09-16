package olion

type Attribute uint16

const (
	ColorDefault Attribute = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

type Part struct {
	dots  []Coordinates
	color Attribute
}

type Parter interface {
	getDots() []Coordinates
	getColor() Attribute
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

func (p Part) getColor() Attribute {
	return p.color
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

func newLinePart(p Parter) LinePart {
	// Todo: check len(p.getDots())
	return LinePart{
		Part: p.(Part),
	}
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
	getPosition() Coordinates
}

func (obj *Object) shape() []Parter {
	return obj.parts
}

func (obj *Object) addPart(p Parter) {
	obj.parts = append(obj.parts, p)
}

func (obj *Object) getPosition() Coordinates {
	return obj.position
}

type Star struct {
	Object
}

func newStar(s int, c Coordinates) *Star {
	star := Star{}
	star.size = s
	star.position = c
	dot := newDotPart(Part{
		dots: []Coordinates{
			Coordinates{X: 0, Y: 0, Z: 0},
		},
		color: ColorWhite,
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
	rectangle1 := newRectanglePart(Part{
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
		},
		color: ColorRed,
	})
	ship.addPart(rectangle1)
	line1 := newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: s, Y: 0, Z: 0},
			Coordinates{X: s / 2, Y: 0, Z: 0},
		},
		color: ColorRed,
	})
	ship.addPart(line1)
	line2 := newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: -s, Y: 0, Z: 0},
			Coordinates{X: -s / 2, Y: 0, Z: 0},
		},
		color: ColorRed,
	})
	ship.addPart(line2)
	line3 := newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: s, Y: s / 2, Z: 0},
			Coordinates{X: s, Y: -s / 2, Z: 0},
		},
		color: ColorRed,
	})
	ship.addPart(line3)
	line4 := newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: -s, Y: s / 2, Z: 0},
			Coordinates{X: -s, Y: -s / 2, Z: 0},
		},
		color: ColorRed,
	})
	ship.addPart(line4)
	return &ship
}
