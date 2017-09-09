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
	dots() []Coordinates
}

func (p *Part) getDots() []Coordinates {
	return p.dots
}

type DotPart struct {
	Part
}

type LinePart struct {
	Part
}

type RectanglePart struct {
	Part
	fill bool
}

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
}

func (obj *Object) shape() []Part {
	return obj.parts
}

type Star struct {
	Object
}

func newStar(s int) *Star {
	return &Star{
		Object{
			size: s,
			parts: []Part{
				Part{
					dots: []Coordinates{
						Coordinates{X: 0, Y: 0, Z: 0},
					},
				},
			},
		},
	}
}

type SpaceShip struct {
	Object
}

func newSpaceShip(s int) *SpaceShip {
	return &SpaceShip{
		Object{
			size: s,
			parts: []Part{
				Part{
					dots: []Coordinates{
						Coordinates{X: s / 2, Y: s / 2, Z: 0},
						Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
					},
				},
			},
		},
	}
}
