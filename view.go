package olion

import (
	"fmt"
	"math"
	"sort"

	"github.com/nsf/termbox-go"
)

type Dot struct {
	X int
	Y int
}

type Screen struct {
	Width     int
	Height    int
	Distance  int
	Vibration int
}

func NewScreen() *Screen {
	w, h := termbox.Size()
	d := 5
	//fmt.Printf("\nW=%v H=%v\n", int(w), int(h))
	return &Screen{Width: int(w), Height: int(h), Distance: d}
}

func (sc *Screen) clear() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func (sc *Screen) flush() {
	termbox.Flush()
}

func (sc *Screen) cover(dot Dot) bool {
	return 0 <= dot.X && dot.X <= sc.Width && 0 <= dot.Y && dot.Y <= sc.Height
}

func (sc *Screen) cover2(dot1, dot2 Dot) bool {
	if dot1.X < 0 && dot2.X < 0 || dot1.X > sc.Width && dot2.X > sc.Width {
		return false
	}
	if dot1.Y < 0 && dot2.Y < 0 || dot1.Y > sc.Height && dot2.Y > sc.Height {
		return false
	}
	return true
}

func (sc *Screen) printString(dot *Dot, str string) {
	//x, y := dot.X, sc.Height-dot.Y+1
	x, y := dot.X, dot.Y
	w, h := termbox.Size()
	for _, r := range []rune(str) {
		idx := y*w + x
		if idx > w*h {
			return
		}
		c := &termbox.CellBuffer()[idx]
		if sc.cover(*dot) {
			if r == '\n' {
				y = y + 1
				x = dot.X - 1
			} else {
				termbox.SetCell(x, y, r, termbox.ColorWhite, c.Bg)
			}
		}
		x += 1
	}
}

func (sc *Screen) printDot(dot *Dot, color Attribute) {
	if sc.cover(*dot) {
		termbox.SetCell(dot.X+sc.Vibration, sc.Height-dot.Y+1, ' ', termbox.ColorDefault, termbox.Attribute(color))
		if sc.Vibration != 0 {
			sc.Vibration = -sc.Vibration
		}
	}
}

func (sc *Screen) printLine(d1, d2 *Dot, color Attribute) {
	if d1 == nil || d2 == nil {
		return
	}
	if !sc.cover2(*d1, *d2) {
		return
	}

	if d1.X == d2.X {
		x := d1.X
		for y := d1.Y; y != d2.Y; {
			sc.printDot(&Dot{X: x, Y: y}, color)
			if d1.Y < d2.Y {
				y++
			} else {
				y--
			}
		}
		return
	}

	orderByX := func(d1, d2 *Dot) (*Dot, *Dot) {
		if d1.X >= d2.X {
			return d2, d1
		} else {
			return d1, d2
		}
	}
	dx1, dx2 := orderByX(d1, d2)
	for x := dx1.X; x <= dx2.X; x++ {
		y1 := dx1.Y + (dx2.Y-dx1.Y)*(x-dx1.X)/(dx2.X-dx1.X)
		y2 := dx1.Y + (dx2.Y-dx1.Y)*(x+1-dx1.X)/(dx2.X-dx1.X)
		if y1 == y2 || x == dx2.X {
			y := y1
			sc.printDot(&Dot{X: x, Y: y}, color)
		} else if y1 < y2 {
			//for y := y1; y < y2 && y != dx1.Y && y != dx2.Y; y++ {
			for y := y1; y < y2; y++ {
				sc.printDot(&Dot{X: x, Y: y}, color)
			}
		} else {
			//for y := y2; y < y1 && y != dx1.Y && y != dx2.Y; y++ {
			for y := y1; y > y2; y-- {
				sc.printDot(&Dot{X: x, Y: y}, color)
			}
		}
	}
}

func (sc *Screen) printRectangle(d1, d2 *Dot, color Attribute, fill bool) {
	//fmt.Printf("d1=%v\td2=%v\n", d1, d2)
	if fill {
		var diffY int
		if d1.Y < d2.Y {
			diffY = 1
		} else {
			diffY = -1
		}
		for y := d1.Y; y != d2.Y; y += diffY {
			sc.printLine(&Dot{X: d1.X, Y: y}, &Dot{X: d2.X, Y: y}, color)
		}
	} else {
		sc.printLine(&Dot{X: d1.X, Y: d1.Y}, &Dot{X: d1.X, Y: d2.Y}, color)
		sc.printLine(&Dot{X: d1.X, Y: d2.Y}, &Dot{X: d2.X, Y: d2.Y}, color)
		sc.printLine(&Dot{X: d2.X, Y: d2.Y}, &Dot{X: d2.X, Y: d1.Y}, color)
		sc.printLine(&Dot{X: d2.X, Y: d1.Y}, &Dot{X: d1.X, Y: d1.Y}, color)
	}
}

func (sc *Screen) printPolygon(dots []Dot, color Attribute, fill bool) {
	if len(dots) < 3 {
		return
	}
	if fill {
		for i := 1; i < len(dots)-1; i++ {
			sc.printTriangle([]Dot{dots[0], dots[i], dots[i+1]}, color)
		}
	} else {
		d1 := dots[0]
		for _, d2 := range dots[1:] {
			sc.printLine(&d1, &d2, color)
			d1 = d2
		}
		sc.printLine(&d1, &dots[0], color)
	}
}

func (sc *Screen) printTriangle(dots []Dot, color Attribute) {
	if len(dots) != 3 {
		return
	}
	getMinMax := func(dots []Dot) (minX, maxX int) {
		minX = dots[0].X
		maxX = dots[0].X
		for i := 1; i < len(dots); i++ {
			if dots[i].X < minX {
				minX = dots[i].X
			} else if dots[i].X > maxX {
				maxX = dots[i].X
			}
		}
		return minX, maxX
	}
	crossPoints := func(dots []Dot, x int) (int, int) {
		var ret []int
		for i := 0; i < len(dots); i++ {
			d1 := dots[i]
			d2 := dots[(i+1)%len(dots)]
			if (d1.X > x && d2.X > x) || (d1.X < x && d2.X < x) || (d1.X == d2.X) {
				//fmt.Printf("d1=%v d2=%v x=%v\n", d1, d2, x)
				continue
			}
			if d1.Y == d2.Y {
				ret = append(ret, d1.Y)
			} else {
				y := d1.Y + (d2.Y-d1.Y)*(d1.X-x)/(d1.X-d2.X)
				ret = append(ret, y)
			}
		}
		if len(ret) < len(dots)-1 {
			//fmt.Printf("\nx=%v len(ret)=%v\n", x, len(ret))
			return 0, 0
		}
		if len(ret) == 2 || ret[0] != ret[1] {
			return ret[0], ret[1]
		} else {
			return ret[0], ret[2]
		}
	}
	minX, maxX := getMinMax(dots)
	debug.Printf("minX=%v maxX=%v\n", minX, maxX)
	for x := minX; x <= maxX; x++ {
		y1, y2 := crossPoints(dots, x)
		//fmt.Printf("x=%v y1=%v y2=%v\n", x, y1, y2)
		sc.printLine(&Dot{X: x, Y: y1}, &Dot{X: x, Y: y2}, color)
	}
}

func (sc *Screen) printCircle(d *Dot, r int, color Attribute, fill bool) {
	for x := d.X - r; x <= d.X+r; x++ {
		//h := int(math.Sqrt(float64(r*r - (x-r)*(x-r))))
		h := int(math.Sqrt(float64(r*r - (x-d.X)*(x-d.X))))
		if fill {
			for y := d.Y - h; y <= d.Y+h; y++ {
				sc.printDot(&Dot{X: x, Y: y}, color)
			}
		} else {
			sc.printDot(&Dot{X: x, Y: d.Y - h}, color)
			sc.printDot(&Dot{X: x, Y: d.Y + h}, color)
		}
	}
}

//https://github.com/sjmudd/ps-top/blob/master/screen/screen.go
// TermBoxChan creates a channel for termbox.Events and run a poller to send
// these events to the channel.  Return the channel to the caller..
func (sc *Screen) TermBoxChan() chan termbox.Event {
	termboxChan := make(chan termbox.Event)
	go func() {
		for {
			termboxChan <- termbox.PollEvent()
		}
	}()
	return termboxChan
}

type View struct {
	state *Olion
}

func NewView(state *Olion) *View {
	return &View{state: state}
}

func drawLine(x, y int, str string) {
	color := termbox.ColorDefault
	runes := []rune(str)

	for i := 0; i < len(runes); i += 1 {
		termbox.SetCell(x+i, y, runes[i], color, 1)
	}
}

func (view *View) mapObject(objPosition Coordinates) *Dot {
	myScreen := view.state.screen
	myPosition := view.state.position
	diffX := objPosition.X - myPosition.X
	diffY := objPosition.Y - myPosition.Y
	diffZ := objPosition.Z - myPosition.Z
	if diffZ <= 0 {
		return nil
	}
	dot := Dot{
		X: int(diffX*myScreen.Distance/diffZ) + myScreen.Width/2,
		Y: int(diffY*myScreen.Distance/diffZ)/2 + myScreen.Height/2, //横に長くする
	}
	/*
		fmt.Printf("mapObject ObjectPosition:%v Screen:%v Position:%v Direction:%v", objPosition, myScreen, myPosition, myDirection)
		fmt.Printf(" sinTheta=%v cosTheta=%v sinPhi=%v cosPhi=%v ", sinTheta, cosTheta, sinPhi, cosPhi)
		fmt.Printf(" diffX:%v diffY:%v diffZ:%v X:%v Y:%v Z:%v", diffX, diffY, diffZ, myCoordinates.X, myCoordinates.Y, myCoordinates.Z)
		fmt.Printf(" map=>%v \n", dot)
	*/
	return &dot
}

func (view *View) mapLength(objPosition Coordinates, length int) int {
	myScreen := view.state.screen
	myPosition := view.state.position
	diffZ := objPosition.Z - myPosition.Z
	//fmt.Printf("objPosition=%v length=%v diffZ=%v \n", objPosition, length, diffZ)
	if diffZ > 0 {
		return int(length * myScreen.Distance / diffZ)
	} else {
		return length * myScreen.Distance
	}
}

func (view *View) draw(upMsgs []upMessage) {
	// Sort Object
	sort.Slice(upMsgs, func(i, j int) bool {
		return upMsgs[i].position.Z > upMsgs[j].position.Z
	})
	//fmt.Printf("\n==>drawObjects(%v)\n", len(view.state.space.Objects))
	for _, msg := range upMsgs {
		position := msg.position
		sort.Slice(msg.parts, func(i, j int) bool {
			avg_dot := func(cs []Coordinates) Coordinates {
				var tx, ty, tz int
				for _, c := range cs {
					tx, ty, tz = tx+position.X+c.X, ty+position.Y+c.Y, tz+position.Z+c.Z
				}
				return Coordinates{X: tx / len(cs), Y: ty / len(cs), Z: tz / len(cs)}
			}
			avg_di := avg_dot(msg.parts[i].getCurDots())
			avg_dj := avg_dot(msg.parts[j].getCurDots())
			//fmt.Printf("avg_di=%v avg_dj=%v\n", avg_di, avg_dj)
			return avg_di.X*avg_di.X+avg_di.Y*avg_di.Y+avg_di.Z*avg_di.Z > avg_dj.X*avg_dj.X+avg_dj.Y*avg_dj.Y+avg_dj.Z*avg_dj.Z
		})
	label1:
		for _, part := range msg.parts {
			//fmt.Printf("shape OK obj=%v\n", obj)
			//fmt.Printf("position=%v\n", obj.getPosition())
			//fmt.Printf("part.getCurDots()=%v\n", part.getCurDots())
			dots := []Dot{}
			for _, dot := range part.getCurDots() {
				d := view.mapObject(Coordinates{
					X: position.X + dot.X,
					Y: position.Y + dot.Y,
					Z: position.Z + dot.Z,
				})
				if d == nil {
					continue label1
				}
				dots = append(dots, *d)
			}
			switch part.(type) {
			case DotPart:
				view.state.screen.printDot(&dots[0], part.getColor())
			case LinePart:
				view.state.screen.printLine(&dots[0], &dots[1], part.getColor())
			case RectanglePart:
				view.state.screen.printRectangle(&dots[0], &dots[1], part.getColor(), part.getFill())
			case CirclePart:
				view.state.screen.printCircle(&dots[0], view.mapLength(position, part.getSize()), part.getColor(), part.getFill())
			case PolygonPart:
				view.state.screen.printPolygon(dots, part.getColor(), part.getFill())
			default:
				fmt.Printf("View.draw -> NO TYPE\n")
			}
		}
	}
}
