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
	getType() Part_type
	getDots() []Coordinates
	addDot(Coordinates)
}

/*
func newPart() Part {
	return Part{}
}
*/

func (p Part) getType() Part_type {
	return p.Type
}

func (p Part) getDots() []Coordinates {
	return p.dots
}

func (p Part) addDot(d Coordinates) {
	p.dots = append(p.dots, d)
}

type DotPart struct {
	Part
}

func newDotPart(p Parter) DotPart {
	// Todo: check len(p.getDots())
	//fmt.Printf("p.(Part)=%v\n", p.(Part))
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
	parts    []Parter
	position Coordinates //位置
	//Direction Direction   //方向
	//Speed
	//created
	//weight

	Type Obj_type
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
	//fmt.Printf(" addPart p=%v  obj.Parts=%v  ==>  ", p, obj.parts)
	obj.parts = append(obj.parts, p)
	//fmt.Printf(" addPart p=%v  obj.Parts=%v ==> ", p, obj.parts)
}

type Star struct {
	Object
}

func newStar(s int, c Coordinates) *Star {
	star := Star{}
	star.size = s
	star.position = c
	star.parts = []Parter{}
	dot := newDotPart(Part{
		Type: Part_Dot,
		dots: []Coordinates{
			Coordinates{X: 0, Y: 0, Z: 0},
		},
	})
	//fmt.Printf("star.parts=%v dot=%v   ==> ", star.parts, dot)
	star.addPart(dot)
	//fmt.Printf("star.parts=%v dot=%v\n", star.parts, dot)
	return &star
}

type SpaceShip struct {
	Object
}

func newSpaceShip(s int, c Coordinates) *SpaceShip {
	ship := SpaceShip{}
	ship.size = s
	ship.position = c
	ship.parts = []Parter{}
	rectangle := newRectanglePart(Part{
		Type: Part_Rectangle,
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
		},
	})
	//fmt.Printf("ship.parts=%v dot=%v   ==> ", ship.parts, rectangle)
	ship.addPart(rectangle)
	//fmt.Printf("ship.parts=%v dot=%v\n", ship.parts, rectangle)
	return &ship
}
