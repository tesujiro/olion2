package olion

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
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

func (sc *Screen) printDot(dot Dot) {
	//fmt.Printf("\x1b[%v;%vH%s", sc.Height-dot.Y+1, dot.X, "X")
	if sc.cover(dot) {
		termbox.SetCell(dot.X, sc.Height-dot.Y+1, ' ', termbox.ColorDefault, 1)
	}
}

func (sc *Screen) printLine(d1, d2 *Dot) {
	if d1 == nil || d2 == nil {
		return
	}
	if (d1.X-d2.X)*(d1.X-d2.X) >= (d1.Y-d2.Y)*(d1.Y-d2.Y) {
		switch {
		case d1.X == d2.X:
			sc.printDot(*d1)
		case d1.X < d2.X:
			for x := d1.X; x <= d2.X; x++ {
				y := d1.Y + (d2.Y-d1.Y)*(x-d1.X)/(d2.X-d1.X)
				sc.printDot(Dot{X: x, Y: y})
			}
		case d1.X > d2.X:
			for x := d2.X; x <= d1.X; x++ {
				y := d2.Y + (d1.Y-d2.Y)*(x-d2.X)/(d1.X-d2.X)
				sc.printDot(Dot{X: x, Y: y})
			}
		}
	} else {
		switch {
		//case d1.Y == d2.Y:
		//sc.printDot(*d1)
		case d1.Y < d2.Y:
			for y := d1.Y; y <= d2.Y; y++ {
				x := d1.X + (d2.X-d1.X)*(y-d1.Y)/(d2.Y-d1.Y)
				sc.printDot(Dot{X: x, Y: y})
			}
		case d1.Y > d2.Y:
			for y := d2.Y; y <= d1.Y; y++ {
				x := d2.X + (d1.X-d2.X)*(y-d2.Y)/(d1.Y-d2.Y)
				sc.printDot(Dot{X: x, Y: y})
			}
		}
	}
}

func (sc *Screen) printCircle(d *Dot, r int, fill bool) {
}

func (sc *Screen) printRectangle(d1, d2 *Dot, fill bool) {
	//Todo:fill
	//fmt.Printf("d1=%v\td2=%v\n", d1, d2)
	sc.printLine(&Dot{X: d1.X, Y: d1.Y}, &Dot{X: d1.X, Y: d2.Y})
	sc.printLine(&Dot{X: d1.X, Y: d2.Y}, &Dot{X: d2.X, Y: d2.Y})
	sc.printLine(&Dot{X: d2.X, Y: d2.Y}, &Dot{X: d2.X, Y: d1.Y})
	sc.printLine(&Dot{X: d2.X, Y: d1.Y}, &Dot{X: d1.X, Y: d1.Y})
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

func (view *View) drawObjects() {
	//fmt.Printf("\n==>drawObjects(%v)\n", len(view.state.space.Objects))

	for _, obj := range view.state.space.Objects {
		//dot := Dot{X: obj.Position.X, Y: obj.Position.Y}
		//fmt.Printf("obj=%v\n", obj)
		if shaper, ok := interface{}(obj).(Shaper); ok {
			//fmt.Printf("cast OK obj=%v\n", obj)
			for _, part := range shaper.shape() {
				//fmt.Printf("shape OK obj=%v\n", obj)
				//switch part.getType() {
				switch part.(type) {
				//case Part_Dot:
				case DotPart:
					//fmt.Printf("Part_Dot: Obj=%v type=%v obj.position=%v\n", obj, reflect.TypeOf(obj), obj.(*Star).position)
					if dot := view.mapObject(obj.(*Star).position); dot != nil {
						view.state.screen.printDot(*dot)
						//fmt.Printf("dot=%v", *dot)
						//view.drawn = append(view.drawn, *dot)
					}
				//case Part_Circle:
				//fmt.Printf("Part_Circle")
				//case Part_Rectangle:
				case RectanglePart:
					position := obj.(*SpaceShip).position
					dots := part.getDots()
					dot1 := view.mapObject(Coordinates{
						X: position.X + dots[0].X,
						Y: position.Y + dots[0].Y,
						Z: position.Z + dots[0].Z,
					})
					dot2 := view.mapObject(Coordinates{
						X: position.X + dots[1].X,
						Y: position.Y + dots[1].Y,
						Z: position.Z + dots[1].Z,
					})
					if dot1 != nil && dot2 != nil {
						//fmt.Printf("d1=%v\td2=%v", dot1, dot2)
						//fmt.Printf("dots=%v\tposition=%v\td1=%v\td2=%v\n", dots, position, dot1, dot2)
						view.state.screen.printRectangle(dot1, dot2, false)
					}
				//fmt.Printf("Part_Part_Rectangle: Obj=%v\n", obj)
				default:
					fmt.Printf("NO TYPE\n")
				}
			}
		} else {

			fmt.Printf("CAST ERROR\n")
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
			view.drawObjects()
			view.state.position = Coordinates{
				X: view.state.position.X, //ここにカーソル移動を入れる
				Y: view.state.position.Y, //ここにカーソル移動を入れる
				Z: view.state.position.Z + view.state.speed,
			}
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
