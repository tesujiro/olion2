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

func (sc *Screen) distance(p1, p2 Coordinates) int {
	return int(math.Sqrt(float64((p1.X-p2.X)*(p1.X-p2.X) + (p1.Y-p2.Y)*(p1.Y-p2.Y) + (p1.Z-p2.Z)*(p1.Z-p2.Z)*(sc.Width*sc.Width))))
	//return int(math.Sqrt(float64((p1.X-p2.X)*(p1.X-p2.X) + (p1.Y-p2.Y)*(p1.Y-p2.Y) + (p1.Z-p2.Z)*(p1.Z-p2.Z)/(sc.Width*sc.Width))))
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

func (sc *Screen) getLinedDots(d1, d2 *Dot) []*Dot {
	var result []*Dot
	if d1 == nil || d2 == nil {
		return result
	}
	if !sc.cover2(*d1, *d2) {
		return result
	}

	if d1.X == d2.X {
		x := d1.X
		if d1.Y < d2.Y {
			for y := d1.Y; y <= d2.Y; y++ {
				result = append(result, &Dot{X: x, Y: y})
			}
		} else {
			for y := d2.Y; y <= d1.Y; y++ {
				result = append(result, &Dot{X: x, Y: y})
			}
		}
		return result
	}

	orderByX := func(d1, d2 *Dot) (*Dot, *Dot) {
		if d1.X >= d2.X {
			return d2, d1
		} else {
			return d1, d2
		}
	}
	dx1, dx2 := orderByX(d1, d2)
	result = make([]*Dot, dx2.X-dx1.X+1)
	for x := dx1.X; x <= dx2.X; x++ {
		y1 := dx1.Y + (dx2.Y-dx1.Y)*(x-dx1.X)/(dx2.X-dx1.X)
		y2 := dx1.Y + (dx2.Y-dx1.Y)*(x+1-dx1.X)/(dx2.X-dx1.X)
		if y1 == y2 || x == dx2.X {
			y := y1
			//result = append(result, &Dot{X: x, Y: y})
			result[x-dx1.X] = &Dot{X: x, Y: y}
		} else if y1 < y2 {
			//for y := y1; y < y2 && y != dx1.Y && y != dx2.Y; y++ {
			for y := y1; y < y2; y++ {
				//result = append(result, &Dot{X: x, Y: y})
				result[x-dx1.X] = &Dot{X: x, Y: y}
			}
		} else {
			//for y := y2; y < y1 && y != dx1.Y && y != dx2.Y; y++ {
			for y := y1; y > y2; y-- {
				//result = append(result, &Dot{X: x, Y: y})
				result[x-dx1.X] = &Dot{X: x, Y: y}
			}
		}
	}
	return result
}

func (sc *Screen) printLine(d1, d2 *Dot, color Attribute) {
	if d1 == nil || d2 == nil {
		return
	}
	/*
		if !sc.cover2(*d1, *d2) {
			return
		}
	*/
	for _, d := range sc.getLinedDots(d1, d2) {
		sc.printDot(d, color)
	}
}

func (sc *Screen) printRectangle(d1, d2 *Dot, color Attribute, fill bool) {
	//debug.Printf("printRectangle fill=%v\td1=%v\td2=%v\t\n", fill, d1, d2)
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
	//debug.Printf("printPolygon fill=%v\tdots=%v\n", fill, dots)
	if len(dots) < 3 {
		return
	}
	if fill {
		for i := 1; i < len(dots)-1; i++ {
			//debug.Printf("printPolygon -> printTriangle\td1=%v\td2=%v\td3=%v\n", dots[0], dots[i], dots[i+1])
			sc.printTriangle([]Dot{dots[0], dots[i], dots[i+1]}, color)
		}
	} else {
		d1 := dots[0]
		for _, d2 := range dots[1:] {
			sc.printLine(&d1, &d2, color)
			//debug.Printf("printPolygon -> printLine\td1=%v\td2=%v\n", d1, d2)
			d1 = d2
		}
		sc.printLine(&d1, &dots[0], color)
		//debug.Printf("printPolygon -> printLine\td1=%v\td2=%v\n", d1, dots[0])
	}
}

func (sc *Screen) printTriangle(dots []Dot, color Attribute) {
	if len(dots) != 3 {
		return
	}
	getLongLineByX := func(dots []Dot) []Dot {
		result := []Dot{dots[0], dots[1], dots[2]}
		sort.Slice(result, func(i, j int) bool {
			return result[i].X <= result[j].X
		})
		return result
	}
	// Todo: performance of loop
	dotsOrderedByX := getLongLineByX(dots)
	for _, d1 := range sc.getLinedDots(&dotsOrderedByX[0], &dotsOrderedByX[2]) {
		if d1.X <= dotsOrderedByX[1].X {
			for _, d2 := range sc.getLinedDots(&dotsOrderedByX[0], &dotsOrderedByX[1]) {
				if d1.X == d2.X {
					sc.printLine(d1, d2, color)
				}
			}
		} else {
			for _, d2 := range sc.getLinedDots(&dotsOrderedByX[1], &dotsOrderedByX[2]) {
				if d1.X == d2.X {
					sc.printLine(d1, d2, color)
				}
			}
		}
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
			case *DotPart:
				view.state.screen.printDot(&dots[0], part.getColor())
			case *LinePart:
				view.state.screen.printLine(&dots[0], &dots[1], part.getColor())
			case *CirclePart: //CirclePartは*Part出なく、Partの埋め込みにしたため型が異なる。
				view.state.screen.printCircle(&dots[0], view.mapLength(position, part.getSize()), part.getColor(), part.getFill())
			case *PolygonPart:
				view.state.screen.printPolygon(dots, part.getColor(), part.getFill())
			default:
				fmt.Printf("View.draw -> NO TYPE\n")
			}
		}
	}
}
