package olion

import (
	"fmt"
	"sort"

	"github.com/nsf/termbox-go"
)

type Dot struct {
	X int
	Y int
}

type Screen struct {
	Width    int
	Height   int
	Distance int
}

func NewScreen() *Screen {
	w, h := termbox.Size()
	d := 5
	fmt.Printf("\nW=%v H=%v\n", int(w), int(h))
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

func (sc *Screen) printDot(dot *Dot, color Attribute) {
	if sc.cover(*dot) {
		termbox.SetCell(dot.X, sc.Height-dot.Y+1, ' ', termbox.ColorDefault, termbox.Attribute(color))
	}
}

func (sc *Screen) printLine(d1, d2 *Dot, color Attribute) {
	if d1 == nil || d2 == nil {
		return
	}
	if !sc.cover2(*d1, *d2) {
		return
	}

	if (d1.X-d2.X)*(d1.X-d2.X) >= (d1.Y-d2.Y)*(d1.Y-d2.Y) {
		switch {
		case d1.X == d2.X:
			sc.printDot(d1, color)
		case d1.X < d2.X:
			for x := d1.X; x <= d2.X; x++ {
				y := d1.Y + (d2.Y-d1.Y)*(x-d1.X)/(d2.X-d1.X)
				sc.printDot(&Dot{X: x, Y: y}, color)
			}
		case d1.X > d2.X:
			for x := d2.X; x <= d1.X; x++ {
				y := d2.Y + (d1.Y-d2.Y)*(x-d2.X)/(d1.X-d2.X)
				sc.printDot(&Dot{X: x, Y: y}, color)
			}
		}
	} else {
		switch {
		case d1.Y < d2.Y:
			for y := d1.Y; y <= d2.Y; y++ {
				x := d1.X + (d2.X-d1.X)*(y-d1.Y)/(d2.Y-d1.Y)
				sc.printDot(&Dot{X: x, Y: y}, color)
			}
		case d1.Y > d2.Y:
			for y := d2.Y; y <= d1.Y; y++ {
				x := d2.X + (d1.X-d2.X)*(y-d2.Y)/(d1.Y-d2.Y)
				sc.printDot(&Dot{X: x, Y: y}, color)
			}
		}
	}
}

func (sc *Screen) printCircle(d *Dot, r int, color Attribute, fill bool) {
}

func (sc *Screen) printRectangle(d1, d2 *Dot, color Attribute, fill bool) {
	//Todo:fill
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

func (sc *Screen) printTriangle(d1, d2, d3 *Dot, fill bool) {
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

//func (view *View) draw(objects []Exister) {
func (view *View) draw(upMsgs []upMessage) {
	// Sort Object
	sort.Slice(upMsgs, func(i, j int) bool {
		return upMsgs[i].position.Z > upMsgs[j].position.Z
	})
	//fmt.Printf("\n==>drawObjects(%v)\n", len(view.state.space.Objects))
	for _, msg := range upMsgs {
		position := msg.position
	label1:
		for _, part := range msg.parts {
			//fmt.Printf("shape OK obj=%v\n", obj)
			//fmt.Printf("position=%v\n", obj.getPosition())
			dots := []Dot{}
			for _, dot := range part.getDots() {
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
				r, _ := part.(RectanglePart)
				view.state.screen.printRectangle(&dots[0], &dots[1], part.getColor(), r.getFill())
			default:
				fmt.Printf("NO TYPE\n")
			}
		}
	}
}
