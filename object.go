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
}

type DotPart struct {
	dot Coordinates
}

type LinePart struct {
	dot1, dot2 Coordinates
}

type RectanglePart struct {
	dot1, dot2 Coordinates
	fill       bool
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
	Position Coordinates //位置
	//Direction Direction   //方向
	Type Obj_type
	Size int
}

type Shaper interface {
	shape() []Part
}

type Star struct {
	Object
}

func (star *Star) shape() []Part {
	return []Part{
		&DotPart{dot: Coordinates{X: 0, Y: 0, Z: 0}},
	}
}

type SpaceShip struct {
	Object
}

func (ship *SpaceShip) shape() []Part {
	dot1 := Coordinates{X: ship.size / 2, Y: ship.size / 2, Z: 0}
	dot2 := Coordinates{X: -ship.size / 2, Y: -ship.size / 2, Z: 0}
	return []Part{
		&RectanglePart{dot1: dot1, dot2: dot2},
		//Part_Line(dot1, dot3),
		//Part_rectangle(dot3, dot4),
	}
}
