package olion

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type Attribute uint16

/*
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
*/

type Part struct {
	curDots []Coordinates
	dots    []Coordinates
	color   Attribute
	size    int
	fill    bool
}

type Parter interface {
	getCurDots() []Coordinates
	setCurDots([]Coordinates)
	getDots() []Coordinates
	getColor() Attribute
	setColor(Attribute)
	getSize() int
	addDot(Coordinates)
	setFill(bool)
	getFill() bool
}

func (p *Part) getDots() []Coordinates {
	return p.dots
}

func (p *Part) getCurDots() []Coordinates {
	if p.curDots == nil {
		return p.dots
	} else {
		//fmt.Printf("\ncurDots=%v\n", p.curDots)
		return p.curDots
	}
}

func (p *Part) setCurDots(cs []Coordinates) {
	p.curDots = cs
}

func (p *Part) addDot(d Coordinates) {
	p.dots = append(p.dots, d)
}

func (p *Part) getColor() Attribute {
	return p.color
}

func (p *Part) setColor(a Attribute) {
	p.color = a
}

func (p *Part) getSize() int {
	return p.size
}

func (p *Part) setFill(b bool) {
	p.fill = b
}

func (p *Part) getFill() bool {
	return p.fill
}

/* one dot part */
type DotPart struct {
	Part
}

type LinePart struct {
	Part
}

/*
func newLinePart(p Parter) LinePart {
	// Todo: check len(p.getDots())
	return LinePart{
		Part: p.(*Part),
	}
}
*/

/*
type RectanglePart struct {
	*Part
}
*/

func newRectanglePart(p Parter) *PolygonPart {
	ds := p.getDots()
	// Todo: check len(p.getDots())
	c0 := ds[0]
	c2 := ds[1]
	var c1, c3 Coordinates
	switch {
	case c0.X == c2.X:
		c1 = Coordinates{X: c0.X, Y: c0.Y, Z: c2.Z}
		c3 = Coordinates{X: c0.X, Y: c2.Y, Z: c0.Z}
	case c0.Y == c2.Y:
		c1 = Coordinates{X: c0.X, Y: c0.Y, Z: c2.Z}
		c3 = Coordinates{X: c2.X, Y: c0.Y, Z: c0.Z}
	case c0.Z == c2.Z:
		c1 = Coordinates{X: c0.X, Y: c2.Y, Z: c0.Z}
		c3 = Coordinates{X: c2.X, Y: c0.Y, Z: c0.Z}
	}

	return &PolygonPart{
		Part{
			dots:  []Coordinates{c0, c1, c2, c3},
			color: p.getColor(),
			size:  p.getSize(),
			fill:  p.getFill(),
		},
	}
}

type PolygonPart struct {
	Part
}

type CirclePart struct {
	Part
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

type Object struct {
	parts []Parter
	size  int
	//weight
	mobile
	position     Coordinates //位置
	prevPosition Coordinates //異動前の位置

	downChannel       downChannel
	upChannel         upChannel
	bomb              bool
	bombable          bool
	throwBombDistance int
	explodedAt        time.Time
	exploding         bool
}

type Exister interface {
	downCh() downChannel
	upCh() upChannel
	run(context.Context, func())
	getPosition() Coordinates
	setPosition(Coordinates)
	getPrevPosition() Coordinates
	setPrevPosition(Coordinates)
	getSpeed() Coordinates
	getSize() int
	setSize(int)
	isBomb() bool
	hasBomb() bool
	removeBomb()
	getThrowBombDistance() int
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

func (obj *Object) getPrevPosition() Coordinates {
	return obj.prevPosition
}

func (obj *Object) setPrevPosition(p Coordinates) {
	obj.prevPosition = p
}

func (obj *Object) getSize() int {
	return obj.size
}

func (obj *Object) setSize(size int) {
	obj.size = size
}

//Todo: getPartsがcurDotsの更新目的になってしまっているので分けるべき。
func (obj *Object) getParts() []Parter {
	return obj.parts
}

func (obj *Object) updateCurDots(currentTime time.Time) {

	spinXY, _ := obj.getSpin() // Todo: spinXZ
	deltaTime := float64(float64(currentTime.Second()) + float64(currentTime.Nanosecond())/1000000000)
	//fmt.Printf("\ndeltaTime=%v\n", deltaTime)
	if spinXY == 0 {
		return
	}
	theta := float64(spinXY) / 360.0 * math.Pi * deltaTime
	//fmt.Printf("theta=%v deltaTime=%v\n", theta, deltaTime)
	sinTheta := math.Sin(theta)
	cosTheta := math.Cos(theta)
	for _, part := range obj.parts {
		cs := []Coordinates{}
		for _, dot := range part.getDots() {
			c := Coordinates{
				X: int(cosTheta*float64(dot.X) - sinTheta*float64(dot.Y)),
				Y: int(sinTheta*float64(dot.X) + cosTheta*float64(dot.Y)),
				Z: dot.Z,
			}
			cs = append(cs, c)
		}
		part.setCurDots(cs)
		//fmt.Printf("\ncs=%v part.curDots=%v\n", cs, part.getCurDots())
	}
}

func (obj *Object) isBomb() bool {
	return obj.bomb
}

func (obj *Object) hasBomb() bool {
	return obj.bombable
}

func (obj *Object) removeBomb() {
	obj.bombable = false
	return
}

func (obj *Object) getThrowBombDistance() int {
	return obj.throwBombDistance
}

//func (obj *Object) throwBomb() {
//	obj.hasBomb = false
//	return
//}

func (obj *Object) run(ctx context.Context, cancel func()) {
	defer cancel()
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case downMsg := <-obj.downChannel:
			obj.updateCurDots(downMsg.time)
			parts := obj.getParts()
			distance := obj.getDistance(downMsg.time)
			newPosition := Coordinates{
				X: obj.position.X - downMsg.deltaPosition.X - distance.X,
				Y: obj.position.Y - downMsg.deltaPosition.Y - distance.Y,
				Z: obj.position.Z - downMsg.deltaPosition.Z - distance.Z,
			}
			obj.prevPosition = obj.position
			obj.position = newPosition
			obj.upChannel <- upMessage{
				position: newPosition,
				parts:    parts,
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
	for _, p := range obj.parts {
		p.setColor(colors.name("Red").Attribute())
	}
	obj.spinXY = 1715
	obj.explodedAt = obj.time
	obj.exploding = true
}

func (obj *Object) newRectangular(start Coordinates, width int, height int, depth int, cols []Attribute) {
	//debug.Printf("newRectangular\n")
	var colors []Attribute
	var m []int
	switch len(cols) {
	case 3:
		m = []int{0, 1, 2, 1, 2, 0}
	case 4:
		m = []int{0, 1, 2, 1, 2, 3}
	default:
		m = make([]int, 6)
		for i := 0; i < len(m); i++ {
			m[i] = i % len(cols)
		}
	}
	colors = make([]Attribute, len(m))
	for i := 0; i < len(m); i++ {
		colors[i] = cols[m[i]]
	}

	addRectangular := func(c0 Coordinates, c1 Coordinates, color Attribute, fill bool) {
		r := newRectanglePart(&Part{
			dots:  []Coordinates{c0, c1},
			color: color,
			fill:  fill,
		})
		obj.addPart(r)
	}
	addRectangular(
		Coordinates{X: start.X, Y: start.Y, Z: start.Z},
		Coordinates{X: start.X + width, Y: start.Y + height, Z: start.Z},
		colors[0], true)
	addRectangular(
		Coordinates{X: start.X, Y: start.Y + height, Z: start.Z},
		Coordinates{X: start.X + width, Y: start.Y + height, Z: start.Z + depth},
		colors[1], true)
	addRectangular(
		Coordinates{X: start.X + width, Y: start.Y, Z: start.Z},
		Coordinates{X: start.X + width, Y: start.Y + height, Z: start.Z + depth},
		colors[2], true)
	addRectangular(
		Coordinates{X: start.X + width, Y: start.Y, Z: start.Z},
		Coordinates{X: start.X + width, Y: start.Y + height, Z: start.Z + depth},
		colors[3], true)
	addRectangular(
		Coordinates{X: start.X, Y: start.Y, Z: start.Z},
		Coordinates{X: start.X, Y: start.Y + height, Z: start.Z + depth},
		colors[4], true)
	addRectangular(
		Coordinates{X: start.X, Y: start.Y, Z: start.Z + depth},
		Coordinates{X: start.X + width, Y: start.Y + height, Z: start.Z + depth},
		colors[5], true)
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
	circle := &CirclePart{
		Part{dots: []Coordinates{Coordinates{X: 0, Y: 0, Z: 0}},
			color: colors.name("White").Attribute(),
			fill:  true,
			size:  rand.Intn(2),
		},
	}
	star.addPart(circle)
	return &star
}

type Bomb struct {
	Object
}

func newBomb(t time.Time, s int, position Coordinates, speed Coordinates) *Bomb {
	bomb := Bomb{Object: *newObject()}
	//bomb.position = position
	bomb.position = Coordinates{X: position.X + speed.X, Y: position.Y + speed.Y, Z: position.Z + speed.Z}
	bomb.time = t
	bomb.speed = Coordinates{X: -speed.X, Y: -speed.Y, Z: -speed.Z}
	bomb.bomb = true
	bomb.size = s
	rectangle1 := newRectanglePart(&Part{
		dots: []Coordinates{
			Coordinates{X: s + position.X, Y: s + position.Y, Z: position.Z},
			Coordinates{X: -s + position.X, Y: -s + position.Y, Z: position.Z},
		},
		color: colors.name("Green").Attribute(),
		fill:  false,
	})
	bomb.addPart(rectangle1)
	return &bomb
}

type EnemyBomb struct {
	Object
}

func newEnemyBomb(t time.Time, s int, position Coordinates, speed Coordinates) *EnemyBomb {
	bomb := EnemyBomb{Object: *newObject()}
	//bomb.position = position
	bomb.position = Coordinates{X: position.X + speed.X, Y: position.Y + speed.Y, Z: position.Z + speed.Z}
	bomb.time = t
	bomb.speed = Coordinates{X: -speed.X, Y: -speed.Y, Z: -speed.Z}
	bomb.bomb = true
	bomb.size = s
	rectangle := newRectanglePart(&Part{
		dots: []Coordinates{
			Coordinates{X: s + position.X, Y: s + position.Y, Z: position.Z},
			Coordinates{X: -s + position.X, Y: -s + position.Y, Z: position.Z},
		},
		color: colors.name("Yellow").Attribute(),
		fill:  false,
	})
	bomb.addPart(rectangle)
	rectangle = newRectanglePart(&Part{
		dots: []Coordinates{
			Coordinates{X: s/2 + position.X, Y: s/2 + position.Y, Z: position.Z},
			Coordinates{X: -s/2 + position.X, Y: -s/2 + position.Y, Z: position.Z},
		},
		color: colors.name("Yellow").Attribute(),
		fill:  false,
	})
	bomb.addPart(rectangle)
	return &bomb
}

type FramedRectangle struct {
	Object
}

func newFramedRectangle(t time.Time, s int, c Coordinates) *FramedRectangle {
	fr := FramedRectangle{Object: *newObject()}
	fr.position = c
	fr.time = t
	fr.speed = Coordinates{X: 0, Y: 0, Z: 0}
	fr.bomb = false
	fr.size = s
	fr.spinXY = 180
	fr.spinXZ = 0
	rectangle1 := newRectanglePart(&Part{
		dots: []Coordinates{
			Coordinates{X: s, Y: s, Z: 0},
			Coordinates{X: -s, Y: -s, Z: 0},
		},
		color: colors.name("Red").Attribute(),
		fill:  true,
	})
	fr.addPart(rectangle1)
	rectangle1 = newRectanglePart(&Part{
		dots: []Coordinates{
			Coordinates{X: s, Y: s, Z: 0},
			Coordinates{X: -s, Y: -s, Z: 0},
		},
		color: colors.name("Black").Attribute(),
		fill:  false,
	})
	fr.addPart(rectangle1)
	return &fr
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
	ship.bombable = true
	ship.throwBombDistance = 2000
	ship.spinXY = rand.Intn(180) - 90
	ship.spinXZ = 0
	rectangle1 := &PolygonPart{
		Part{
			dots: []Coordinates{
				Coordinates{X: s / 2, Y: s / 2, Z: 0},
				Coordinates{X: s / 2, Y: -s / 2, Z: 0},
				Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
				Coordinates{X: -s / 2, Y: s / 2, Z: 0},
			},
			color: colors.name("Red").Attribute(),
			fill:  true,
		}}
	ship.addPart(rectangle1)
	rectangle2 := &PolygonPart{
		Part{
			dots: []Coordinates{
				Coordinates{X: s / 2, Y: s / 2, Z: -1},
				Coordinates{X: s / 2, Y: -s / 2, Z: -1},
				Coordinates{X: -s / 2, Y: -s / 2, Z: -1},
				Coordinates{X: -s / 2, Y: s / 2, Z: -1},
			},
			color: colors.name("Black").Attribute(),
			fill:  false,
		}}
	ship.addPart(rectangle2)
	line1 := &LinePart{
		Part{
			dots: []Coordinates{
				Coordinates{X: s, Y: 0, Z: 0},
				Coordinates{X: s / 2, Y: 0, Z: 0},
			},
			color: colors.name("Black").Attribute(),
		}}
	ship.addPart(line1)
	line2 := &LinePart{
		Part{
			dots: []Coordinates{
				Coordinates{X: -s, Y: 0, Z: 0},
				Coordinates{X: -s / 2, Y: 0, Z: 0},
			},
			color: colors.name("Black").Attribute(),
		}}
	ship.addPart(line2)

	line3 := &LinePart{
		Part{
			dots: []Coordinates{
				Coordinates{X: s, Y: s / 2, Z: 0},
				Coordinates{X: s, Y: -s / 2, Z: 0},
			},
			color: colors.name("Black").Attribute(),
		}}
	ship.addPart(line3)
	line4 := &LinePart{
		Part{
			dots: []Coordinates{
				Coordinates{X: -s, Y: s / 2, Z: 0},
				Coordinates{X: -s, Y: -s / 2, Z: 0},
			},
			color: colors.name("Black").Attribute(),
		}}
	ship.addPart(line4)

	var line *LinePart
	line = &LinePart{
		Part{
			dots: []Coordinates{
				Coordinates{X: s, Y: s / 2, Z: 0},
				Coordinates{X: s / 2, Y: s, Z: 0},
			},
			color: colors.name("Black").Attribute(),
		}}
	ship.addPart(line)
	line = &LinePart{
		Part{
			dots: []Coordinates{
				Coordinates{X: s, Y: -s / 2, Z: 0},
				Coordinates{X: s / 2, Y: -s, Z: 0},
			},
			color: colors.name("Black").Attribute(),
		}}
	ship.addPart(line)
	line = &LinePart{
		Part{
			dots: []Coordinates{
				Coordinates{X: -s, Y: s / 2, Z: 0},
				Coordinates{X: -s / 2, Y: s, Z: 0},
			},
			color: colors.name("Black").Attribute(),
		}}
	ship.addPart(line)
	line = &LinePart{
		Part{
			dots: []Coordinates{
				Coordinates{X: -s, Y: -s / 2, Z: 0},
				Coordinates{X: -s / 2, Y: -s, Z: 0},
			},
			color: colors.name("Black").Attribute(),
		}}
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
	ship.speed = Coordinates{
		X: rand.Intn(40) - 20,
		Y: rand.Intn(40) - 20,
		Z: rand.Intn(40),
	}
	height := -50
	ship.addPart(&PolygonPart{
		Part{
			dots: []Coordinates{
				Coordinates{X: s / 2, Y: s / 2, Z: 0},
				Coordinates{X: s / 2, Y: -s / 2, Z: 0},
				Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
				Coordinates{X: -s / 2, Y: s / 2, Z: 0},
			},
			color: colors.name("Black").Attribute(),
			fill:  true,
		}})
	ship.addPart(&PolygonPart{
		Part{
			dots: []Coordinates{
				Coordinates{X: s / 2, Y: s / 2, Z: 0},
				Coordinates{X: s / 2, Y: -s / 2, Z: 0},
				Coordinates{X: 0, Y: 0, Z: height},
			},
			color: colors.name("Red").Attribute(),
			fill:  true,
		}})
	ship.addPart(&PolygonPart{
		Part{
			dots: []Coordinates{
				Coordinates{X: -s / 2, Y: -s / 2, Z: 0},
				Coordinates{X: -s / 2, Y: s / 2, Z: 0},
				Coordinates{X: 0, Y: 0, Z: height},
			},
			color: colors.name("Green").Attribute(),
			fill:  true,
		}})

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

	layers := rand.Intn(5) + 2
	distance := 30
	diff_size := -80
	colors := []Attribute{colors.name("Black").Attribute(), colors.name("Red").Attribute()}
	for i := 0; i < layers; i++ {
		edge_size := s + diff_size*i
		rectangle := newRectanglePart(&Part{
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

type SpaceBox3 struct {
	Object
}

func newBox3(t time.Time, s int, c Coordinates) *SpaceBox3 {
	box := SpaceBox3{Object: *newObject()}
	box.size = s
	box.position = c
	box.time = t
	box.speed = Coordinates{
		X: rand.Intn(10) - 5,
		Y: rand.Intn(10) - 5,
		Z: rand.Intn(10),
	}
	cs := []Attribute{
		colors.name("Black").Attribute(),
		colors.name("Red").Attribute(),
		colors.name("White").Attribute(),
		colors.name("Green").Attribute(),
		colors.name("Yellow").Attribute(),
	}
	box.newRectangular(Coordinates{X: 0, Y: 0, Z: 0}, s, s, s/10, cs)

	return &box
}
