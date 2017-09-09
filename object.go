package olion

type Part_type int

const (
	Part_Dot Part_type = iota
	Part_Line
	Part_Circle
	Part_Rectangle
)

type Part struct {
	Type Part_type
	dots []Coordinates
}

type Parter interface {
	getParts() []Coordinates
}

func (p Part) getDots() []Coordinates {
	return p.dots
}

type DotPart struct {
	Part
}

func newDotPart(p Part) DotPart {
	return DotPart{
		Part: p,
	}
}

//func (p *DotPart) getDots() []Coordinates {
//return p.dots
//}

type LinePart struct {
	Part
}

type RectanglePart struct {
	Part
	fill bool
}

func newRectanglePart(p Part, f bool) DotPart {
	return DotPart{
		Part: p,
		fill: f,
	}
}

//func (p *RectanglePart) getDots() []Coordinates {
//return p.dots
//}

type Obj_type int

const (
	Obj_Dot Obj_type = iota
	//Obj_Line
	Obj_Box
	//Obj_Char
	Obj_Star
)

type Object struct {
	parts    []Part
	position Coordinates //位置
	//Direction Direction   //方向
	//Speed
	//created
	//weight

	Type Obj_type
	size int
}

type Shaper interface {
	shape() []Part
	addPart(Parter)
}

func (obj Object) shape() []Part {
	return obj.parts
}

func (obj Object) addPart(p Parter) {
	obj.parts + append(obj.part, p)
}

type Star struct {
	Object
}

func newStar(s int, c Coordinates) *Star {
	star := Star{}
	star.size = s
	star.position = c
	//part := DotPart{}
	part := Part{}
	part.Type = Part_Dot
	part.dots = []Coordinates{
		Coordinates{X: 0, Y: 0, Z: 0},
	}
	star.parts = []Part{
		part,
		//interface{}(part).(Part),
	}
	star.addPart(newDotPart(part))

	return &star
}

type SpaceShip struct {
	Object
}

func newSpaceShip(s int, c Coordinates) *SpaceShip {
	ship := SpaceShip{}
	ship.size = s
	ship.position = c
	ship.parts = []Part{
		Part{
			Type: Part_Rectangle,
			dots: []Coordinates{
				Coordinates{X: s / 2, Y: s / 2, Z: 0},
				Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
			},
		},
	}
	return &ship
}
