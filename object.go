package olion

import (
	"context"
	"math/rand"
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
	size  int
	fill  bool
}

type Parter interface {
	getDots() []Coordinates
	getColor() Attribute
	getSize() int
	addDot(Coordinates)
	setFill(bool)
	getFill() bool
}

func (p Part) getDots() []Coordinates {
	return p.dots
}

func (p Part) addDot(d Coordinates) {
	p.dots = append(p.dots, d)
}

func (p Part) getColor() Attribute {
	return p.color
}

func (p Part) getSize() int {
	return p.size
}

func (p Part) setFill(b bool) {
	p.fill = b
}

func (p Part) getFill() bool {
	return p.fill
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
}

func newRectanglePart(p Parter) RectanglePart {
	// Todo: check len(p.getDots())
	return RectanglePart{
		Part: p.(Part),
	}
}

type CirclePart struct {
	Part
}

func newCirclePart(p Parter) CirclePart {
	// Todo: check len(p.getDots())
	return CirclePart{
		Part: p.(Part),
	}
}

/*
type Shaper interface {
	shape() []Parter
	addPart(Parter)
	getPosition() Coordinates
	setPosition(Coordinates)
}
*/

type Exister interface {
	downCh() downChannel
	upCh() upChannel
	//quitCh() quitChannel
	run(context.Context, func())
}

type downChannel chan downMessage // Read from Main Loop

type upChannel chan upMessage // Write to Main Loop

type downMessage struct {
	time          time.Time
	deltaPosition Coordinates
}

type upMessage struct {
	position Coordinates
	parts    []Parter
}

type Object struct {
	parts []Parter
	size  int
	//weight
	speed Coordinates
	time  time.Time
	//Direction Direction   //方向
	position Coordinates //位置

	downChannel downChannel
	upChannel   upChannel
}

func newObject() *Object {
	return &Object{
		downChannel: make(downChannel),
		upChannel:   make(upChannel),
	}
}

/*
func (obj *Object) shape() []Parter {
	return obj.parts
}
*/

func (obj *Object) addPart(p Parter) {
	obj.parts = append(obj.parts, p)
}

/*
func (obj *Object) setTime(t time.Time) {
	obj.time = t
}

func (obj *Object) getTime() time.Time {
	return obj.time
}
*/

func (obj *Object) downCh() downChannel {
	return obj.downChannel
}

func (obj *Object) upCh() upChannel {
	return obj.upChannel
}

func (obj *Object) run(ctx context.Context, cancel func()) {
	defer cancel()
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case downMsg := <-obj.downChannel:
			//fmt.Printf("Object: <-obj.downChannel %v\n", downMsg)
			//fmt.Printf("Object: <-obj.downChannel %v\n", downMsg)
			deltaTime := int(downMsg.time.Sub(obj.time) / time.Second)
			/*
				if deltaTime != 0 && obj.speed.X != 0 {
					fmt.Printf("(%d,%d,%d)\n", obj.speed.X*deltaTime, obj.speed.Y*deltaTime, obj.speed.Z*deltaTime)
				}
			*/

			newPosition := Coordinates{
				X: obj.position.X - downMsg.deltaPosition.X - obj.speed.X*deltaTime,
				Y: obj.position.Y - downMsg.deltaPosition.Y - obj.speed.Y*deltaTime,
				Z: obj.position.Z - downMsg.deltaPosition.Z - obj.speed.Z*deltaTime,
			}
			obj.position = newPosition
			obj.time = obj.time.Add(time.Duration(deltaTime) * time.Second)
			obj.upChannel <- upMessage{
				position: newPosition,
				parts:    obj.parts,
			}
		}
	}
}

type Star struct {
	Object
}

func newStar(t time.Time, s int, c Coordinates) *Star {
	star := Star{Object: *newObject()}
	star.size = s
	star.position = c
	star.time = t
	star.speed = Coordinates{
		X: 0,
		Y: 0,
		Z: 0,
	}
	//dot := newDotPart(Part{
	circle := newCirclePart(Part{
		dots: []Coordinates{
			Coordinates{X: 0, Y: 0, Z: 0},
		},
		color: ColorWhite,
		//color: ColorYellow,
		fill: true,
		size: rand.Intn(2),
	})
	star.addPart(circle)
	return &star
}

type SpaceShip struct {
	Object
}

func newSpaceShip(t time.Time, s int, c Coordinates) *SpaceShip {
	ship := SpaceShip{Object: *newObject()}
	ship.size = s
	ship.position = c
	ship.time = t
	ship.speed = Coordinates{
		X: rand.Intn(4) - 2,
		Y: rand.Intn(4) - 2,
		Z: rand.Intn(20),
	}
	rectangle1 := newRectanglePart(Part{
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
		},
		color: ColorRed,
		fill:  true,
	})
	ship.addPart(rectangle1)
	rectangle2 := newRectanglePart(Part{
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
		},
		color: ColorBlack,
		fill:  false,
	})
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
