package olion

import (
	"context"
	"math"
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

type PolygonPart struct {
	Part
}

func newPolygonPart(p Parter) PolygonPart {
	// Todo: check len(p.getDots())
	return PolygonPart{
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

type mobile struct {
	speed  Coordinates
	spinXY int // XY spin speed in radian
	spinXZ int // XZ spin speed in radian
	time   time.Time
}

func (obj *mobile) setTime(t time.Time) {
	obj.time = t
}

func (obj *mobile) getTime() time.Time {
	return obj.time
}

func (obj *mobile) getSpeed() Coordinates {
	return obj.speed
}

func (obj *mobile) getSpin() (int, int) {
	return obj.spinXY, obj.spinXZ
}

func (obj *mobile) getDistance(currentTime time.Time) Coordinates {
	prevTime := obj.getTime()
	speed := obj.getSpeed()
	deltaTime := float64(currentTime.Sub(prevTime) / time.Millisecond)
	obj.setTime(prevTime.Add(time.Duration(deltaTime) * time.Millisecond))
	distance := Coordinates{
		X: int(float64(speed.X) * deltaTime / 100),
		Y: int(float64(speed.Y) * deltaTime / 100),
		Z: int(float64(speed.Z) * deltaTime / 100),
	}
	return distance
}

func (obj *Object) getParts(currentTime time.Time) []Parter {
	prevTime := obj.getTime()
	spinXY, _ := obj.getSpin() // Todo: spinXZ
	deltaTime := float64(currentTime.Sub(prevTime) / time.Millisecond)
	obj.setTime(prevTime.Add(time.Duration(deltaTime) * time.Millisecond))
	if spinXY == 0 {
		return obj.parts
	}
	theta := float64(spinXY) / 360.0 * math.Pi * deltaTime / 100
	sinTheta := math.Sin(theta)
	cosTheta := math.Cos(theta)
	var ret []Parter
	for _, part := range obj.parts {
		cs := []Coordinates{}
		for _, dot := range part.getDots() {
			c := Coordinates{
				X: cosTheta*dot.X - sinTheta*dot.Y,
				Y: sinTheta*dot.X + cosTheta*dot.Y,
				Z: dot.Z,
			}
			cs = append(cs, c)
		}
		p := Part{
			dots:  cs,
			color: part.color,
			size:  part.size,
			fill:  part.fill,
		}
		ret = append(ret, p)
	}
	return ret
}

type Object struct {
	parts []Parter
	size  int
	//weight
	mobile
	position Coordinates //位置

	downChannel downChannel
	upChannel   upChannel
	bomb        bool
	explodedAt  time.Time
	exploding   bool
}

type Exister interface {
	downCh() downChannel
	upCh() upChannel
	run(context.Context, func())
	getPosition() Coordinates
	setPosition(Coordinates)
	getSize() int
	setSize(int)
	isBomb() bool
	explode()
	isExploding() bool
	getExplodedTime() time.Time
}

func newObject() *Object {
	return &Object{
		downChannel: make(downChannel),
		upChannel:   make(upChannel),
	}
}

func (obj *Object) addPart(p Parter) {
	obj.parts = append(obj.parts, p)
}

func (obj *Object) downCh() downChannel {
	return obj.downChannel
}

func (obj *Object) upCh() upChannel {
	return obj.upChannel
}

func (obj *Object) getPosition() Coordinates {
	return obj.position
}

func (obj *Object) setPosition(p Coordinates) {
	obj.position = p
}

func (obj *Object) getSize() int {
	return obj.size
}

func (obj *Object) setSize(size int) {
	obj.size = size
}

func (obj *Object) isBomb() bool {
	return obj.bomb
}

func (obj *Object) run(ctx context.Context, cancel func()) {
	defer cancel()
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case downMsg := <-obj.downChannel:
			distance := obj.getDistance(downMsg.time)
			newPosition := Coordinates{
				X: obj.position.X - downMsg.deltaPosition.X - distance.X,
				Y: obj.position.Y - downMsg.deltaPosition.Y - distance.Y,
				Z: obj.position.Z - downMsg.deltaPosition.Z - distance.Z,
			}
			obj.position = newPosition
			obj.upChannel <- upMessage{
				position: newPosition,
				//parts:    obj.parts,
				parts: obj.getParts(downMsg.time),
			}
		}
	}
}

func (obj *Object) isExploding() bool {
	return obj.exploding
}

func (obj *Object) getExplodedTime() time.Time {
	return obj.explodedAt
}

func (obj *Object) explode() {
	obj.parts = []Parter{}
	circle := newCirclePart(Part{
		dots: []Coordinates{
			Coordinates{X: 0, Y: 0, Z: 0},
		},
		fill:  false,
		color: ColorRed,
		size:  100,
	})
	obj.addPart(circle)
	obj.explodedAt = obj.time
	obj.exploding = true
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

type Bomb struct {
	Object
}

func newBomb(t time.Time, s int, speed Coordinates) *Bomb {
	bomb := Bomb{Object: *newObject()}
	//bomb.position = Coordinates{X: 0, Y: 0, Z: 0}
	bomb.position = speed
	bomb.time = t
	bomb.speed = Coordinates{X: -speed.X, Y: -speed.Y, Z: -speed.Z - 80}
	bomb.bomb = true
	bomb.size = s
	//fmt.Printf("size=%v \n", bomb.getSize())
	rectangle1 := newRectanglePart(Part{
		dots: []Coordinates{
			Coordinates{X: s, Y: s, Z: 0},
			Coordinates{X: -s, Y: -s, Z: 0},
		},
		color: ColorGreen,
		fill:  false,
	})
	bomb.addPart(rectangle1)
	return &bomb
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
		X: rand.Intn(40) - 20,
		Y: rand.Intn(40) - 20,
		Z: rand.Intn(40),
	}
	ship.spinXY = 360
	ship.spinXZ = 0
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

type SpaceBox struct {
	Object
}

func newBox(t time.Time, s int, c Coordinates) *SpaceBox {
	ship := SpaceBox{Object: *newObject()}
	ship.size = s
	ship.position = c
	ship.time = t
	length := s / 20
	flont_size := s / 2
	ship.speed = Coordinates{
		X: rand.Intn(40) - 20,
		Y: rand.Intn(40) - 20,
		Z: rand.Intn(40),
	}
	rectangle1 := newRectanglePart(Part{
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: s / 2, Z: 0},
			Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
		},
		color: ColorBlack,
		fill:  true,
	})
	ship.addPart(rectangle1)

	var line LinePart
	line = newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: s / 2, Z: 0},
			Coordinates{X: flont_size / 2, Y: flont_size / 2, Z: -length},
		},
		color: ColorBlack,
	})
	ship.addPart(line)
	line = newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: -s / 2, Y: s / 2, Z: 0},
			Coordinates{X: -flont_size / 2, Y: flont_size / 2, Z: -length},
		},
		color: ColorBlack,
	})
	ship.addPart(line)
	line = newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
			Coordinates{X: -flont_size / 2, Y: -flont_size / 2, Z: -length},
		},
		color: ColorBlack,
	})
	ship.addPart(line)
	line = newLinePart(Part{
		dots: []Coordinates{
			Coordinates{X: s / 2, Y: -s / 2, Z: 0},
			Coordinates{X: flont_size / 2, Y: -flont_size / 2, Z: -length},
		},
		color: ColorBlack,
	})
	ship.addPart(line)

	rectangle2 := newRectanglePart(Part{
		dots: []Coordinates{
			Coordinates{X: flont_size / 2, Y: flont_size / 2, Z: -length},
			Coordinates{X: -flont_size / 2, Y: -flont_size / 2, Z: -length},
		},
		color: ColorRed,
		fill:  true,
	})
	ship.addPart(rectangle2)

	return &ship
}

type SpaceBox2 struct {
	Object
}

func newBox2(t time.Time, s int, c Coordinates) *SpaceBox2 {
	ship := SpaceBox2{Object: *newObject()}
	ship.size = s
	ship.position = c
	ship.time = t
	ship.speed = Coordinates{
		X: rand.Intn(10) - 5,
		Y: rand.Intn(10) - 5,
		Z: rand.Intn(10),
	}

	layers := 5
	distance := 30
	diff_size := -80
	colors := []Attribute{ColorBlack, ColorRed}
	for i := 0; i < layers; i++ {
		edge_size := s + diff_size*i
		rectangle := newRectanglePart(Part{
			dots: []Coordinates{
				Coordinates{X: edge_size / 2, Y: edge_size / 2, Z: -distance * i},
				Coordinates{X: -edge_size / 2, Y: -edge_size / 2, Z: -distance * i},
			},
			color: colors[i%len(colors)],
			fill:  true,
		})
		ship.addPart(rectangle)
	}

	return &ship
}
