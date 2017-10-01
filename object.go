package olion

import (
	"context"
	"time"
)

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

func (p *RectanglePart) setFill(b bool) {
	p.fill = b
}

func (p *RectanglePart) getFill() bool {
	return p.fill
}

type Object struct {
	parts []Parter
	size  int
	//weight
	speed float32
	time  time.Time
	//Direction Direction   //方向
	position Coordinates //位置
}

type Shaper interface {
	shape() []Parter
	addPart(Parter)
	getPosition() Coordinates
	setPosition(Coordinates)
	//getTime() time.Time
	//setTime(time.Time)
}

type Runner interface {
	run()
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

func (obj *Object) setPosition(c Coordinates) {
	obj.position = c
}

func (obj *Object) setTime(t time.Time) {
	obj.time = t
}

func (obj *Object) getTime() time.Time {
	return obj.time
}

// チャネル
// Main -> Object : Move{X,Y,Z},time
// Object -> Main : SHaper ?  Position,Parts

//func (obj *Object) run(ctx context.Context, cancel func(), inChan Direction, outChan Shaper) {
func (obj *Object) run(ctx context.Context, cancel func()) {
	defer cancel()
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
			//case in := <- inChan
		}
	}
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
	//star.setCreatedTime()
	return &star
}

type SpaceShip struct {
	Object
}

func newSpaceShip(s int, c Coordinates, t time.Time) *SpaceShip {
	ship := SpaceShip{}
	ship.size = s
	ship.position = c
	ship.time = t
	rectangle1 := newRectanglePart(Part{
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
		},
		color: ColorRed,
	})
	rectangle1.setFill(true)
	ship.addPart(rectangle1)
	rectangle2 := newRectanglePart(Part{
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
		},
		color: ColorBlack,
	})
	rectangle2.setFill(false)
	ship.addPart(rectangle2)
	line1 := newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: s, Y: 0, Z: 0},
			Coordinates{X: s / 2, Y: 0, Z: 0},
		},
		color: ColorBlack,
	})

	ship.addPart(line1)
	line2 := newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: -s, Y: 0, Z: 0},
			Coordinates{X: -s / 2, Y: 0, Z: 0},
		},
		color: ColorBlack,
	})
	ship.addPart(line2)

	line3 := newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: s, Y: s / 2, Z: 0},
			Coordinates{X: s, Y: -s / 2, Z: 0},
		},
		color: ColorBlack,
	})
	ship.addPart(line3)
	line4 := newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: -s, Y: s / 2, Z: 0},
			Coordinates{X: -s, Y: -s / 2, Z: 0},
		},
		color: ColorBlack,
	})
	ship.addPart(line4)

	var line LinePart
	line = newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: s, Y: s / 2, Z: 0},
			Coordinates{X: s / 2, Y: s, Z: 0},
		},
		color: ColorBlack,
	})
	ship.addPart(line)
	line = newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: s, Y: -s / 2, Z: 0},
			Coordinates{X: s / 2, Y: -s, Z: 0},
		},
		color: ColorBlack,
	})
	ship.addPart(line)
	line = newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: -s, Y: s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: s, Z: 0},
		},
		color: ColorBlack,
	})
	ship.addPart(line)
	line = newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: -s, Y: -s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: -s, Z: 0},
		},
		color: ColorBlack,
	})
	ship.addPart(line)
	//ship.setCreatedTime()

	return &ship
}
