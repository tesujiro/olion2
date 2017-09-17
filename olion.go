package olion

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"

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
	//fmt.Printf("Color=%v\n", color)
	if sc.cover(*dot) {
		//termbox.SetCell(dot.X, sc.Height-dot.Y+1, ' ', termbox.Attribute(color), 1)
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
	/*
		var diffX, diffY int
		if d1.X < d2.X {
			diffX = 1
		} else {
			diffX = -1
		}
		if d1.Y < d2.Y {
			diffY = 1
		} else {
			diffY = -1
		}
		for x, y := d1.X, d1.Y; x != d2.X && y != d2.Y; x, y = x+diffX, y+diffY {
			sc.printDot(&Dot{X: x, Y: y}, color)
		}
	*/

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
	//drawn []Dot
}

func NewView(state *Olion) *View {
	return &View{state: state}
}

func drawLine(x, y int, str string) {
	color := termbox.ColorDefault
	//backgroundColor := termbox.ColorDefault
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
	//if diffX <= 0 || diffY < +0 || diffZ <= 0 {
	if diffZ <= 0 {
		return nil
	}
	dot := Dot{
		X: int(diffX*myScreen.Distance/diffZ) + myScreen.Width/2,
		Y: int(diffY*myScreen.Distance/diffZ) + myScreen.Height/2,
	}
	/*
		fmt.Printf("mapObject ObjectPosition:%v Screen:%v Position:%v Direction:%v", objPosition, myScreen, myPosition, myDirection)
		fmt.Printf(" sinTheta=%v cosTheta=%v sinPhi=%v cosPhi=%v ", sinTheta, cosTheta, sinPhi, cosPhi)
		fmt.Printf(" diffX:%v diffY:%v diffZ:%v X:%v Y:%v Z:%v", diffX, diffY, diffZ, myCoordinates.X, myCoordinates.Y, myCoordinates.Z)
		fmt.Printf(" map=>%v \n", dot)
	*/
	return &dot
}

func (view *View) move(moveDiff Coordinates) {
	for _, obj := range view.state.space.Objects {
		//Each Object Move
		position := obj.getPosition()
		newPosition := Coordinates{
			X: position.X - moveDiff.X,
			Y: position.Y - moveDiff.Y,
			Z: position.Z - moveDiff.Z,
		}
		obj.setPosition(newPosition)
	}
}

func (view *View) drawObjects() {
	// Sort Object
	sort.Slice(view.state.space.Objects, func(i, j int) bool {
		return view.state.space.Objects[i].getPosition().Z > view.state.space.Objects[j].getPosition().Z
	})

	//fmt.Printf("\n==>drawObjects(%v)\n", len(view.state.space.Objects))
	for _, obj := range view.state.space.Objects {
	label1:
		for _, part := range obj.shape() {
			//fmt.Printf("shape OK obj=%v\n", obj)
			//fmt.Printf("position=%v\n", obj.getPosition())
			position := obj.getPosition()
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

func (view *View) Loop(ctx context.Context, cancel func()) error {
	//func (view *View) Loop(ctx context.Context, cancel func()) {
	defer cancel()
	//fmt.Println("==>Loop")

	TermBoxChan := view.state.screen.TermBoxChan()
	tick := time.NewTicker(time.Millisecond * time.Duration(1)).C
	count := 0
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			view.state.screen.clear()
			view.move(Coordinates{
				X: 0, //ここにカーソル移動を入れる
				Y: 0, //ここにカーソル移動を入れる
				Z: view.state.speed,
			})
			view.drawObjects()
			count++
			drawLine(0, 0, fmt.Sprintf("counter=%v position=%v", count, view.state.position))
			view.state.screen.flush()
		case ev := <-TermBoxChan:
			if ev.Type == termbox.EventKey {
				if ev.Key == termbox.KeyEsc {
					break mainloop // Esc で実行終了
				}
			}
		}
	}
	return nil
}

type Coordinates struct {
	X int
	Y int
	Z int
}

/*
type Direction struct {
	theta float64
	phi   float64
}
*/

type Space struct {
	//Objects []Object
	Objects []Shaper
}

func (spc *Space) addObj(obj Shaper) {
	//spc.Objects = append(spc.Objects, obj.(Shaper))
	spc.Objects = append(spc.Objects, obj)
}

func NewSpace() *Space {
	//fmt.Printf("NewSpace Start")
	spc := &Space{}

	//Add Star
	count := 1000
	w, h := termbox.Size()
	max := int((w + h) * 10)
	min := -max
	depth := (w + h) * 100
	for i := 0; i < count; i++ {
		spc.addObj(
			newStar(1, Coordinates{
				X: (min + rand.Intn(max-min)) * 2,
				Y: min + rand.Intn(max-min),
				Z: rand.Intn(depth),
			}))
	}

	//Add SpaceShip
	count = 100
	for i := 0; i < count; i++ {
		spc.addObj(
			newSpaceShip(rand.Intn(max), Coordinates{
				X: (min + rand.Intn(max-min)) * 2,
				Y: min + rand.Intn(max-min),
				Z: rand.Intn(depth),
			}))
	}

	fmt.Printf("==> %v Objects\n", len(spc.Objects))
	return spc
}

type Olion struct {
	Argv   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	//hub    MessageHub

	//bufferSize int
	// Config contains the values read in from config file
	//config Config
	//currentLineBuffer Buffer
	//maxScanBufferSize int
	readyCh chan struct{}
	screen  *Screen
	space   *Space

	position Coordinates
	speed    int

	// cancelFunc is called for Exit()
	cancelFunc func()
	// Errors are stored here
	err error
}

func New() *Olion {
	rand.Seed(time.Now().UnixNano())

	return &Olion{
		Argv:   os.Args,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		//currentLineBuffer: NewMemoryBuffer(), // XXX revisit this
		readyCh: make(chan struct{}),
		screen:  NewScreen(),
		space:   NewSpace(),
		//maxScanBufferSize: bufio.MaxScanTokenSize,
		position: Coordinates{X: 0, Y: 0, Z: 0},
		speed:    1,
		//cancelFunc: func() {},
	}
}

func (state *Olion) Run(ctx context.Context) (err error) {

	var _cancelOnce sync.Once
	var _cancel func()
	ctx, _cancel = context.WithCancel(ctx)
	cancel := func() {
		_cancelOnce.Do(func() {
			fmt.Printf("Olion.Run cancel called")
			_cancel()
		})
	}

	state.cancelFunc = cancel
	go NewView(state).Loop(ctx, cancel)
	//time.Sleep(5 * time.Second)

	// Alright, done everything we need to do automatically. We'll let
	// the user play with peco, and when we receive notification to
	// bail out, the context should be canceled appropriately
	<-ctx.Done()

	return nil
}
