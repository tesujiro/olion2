package olion

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
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

/*
var cache_sin = map[float64]float64{}
var cache_cos = map[float64]float64{}

func sin(f float64) float64 {
	if c, ok := cache_sin[f]; ok {
		return c
	}
	sin := math.Sin(f)
	cache_sin[f] = sin
	return sin
}

func cos(f float64) float64 {
	if c, ok := cache_cos[f]; ok {
		return c
	}
	cos := math.Cos(f)
	cache_cos[f] = cos
	return cos
}
*/

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
		//fmt.Printf("obj=%v", obj)
		switch obj.Type {
		case Obj_Dot:
			if dot := view.mapObject(obj.Position); dot != nil {
				view.state.screen.printDot(*dot)
				//fmt.Printf("dot=%v", *dot)
				//view.drawn = append(view.drawn, *dot)
			}
		case Obj_Box:
			dot1 := view.mapObject(Coordinates{
				X: obj.Position.X + obj.Size*2,
				Y: obj.Position.Y + obj.Size,
				Z: obj.Position.Z,
			})
			dot2 := view.mapObject(Coordinates{
				X: obj.Position.X + obj.Size*2,
				Y: obj.Position.Y - obj.Size,
				Z: obj.Position.Z,
			})
			dot3 := view.mapObject(Coordinates{
				X: obj.Position.X - obj.Size*2,
				Y: obj.Position.Y - obj.Size,
				Z: obj.Position.Z,
			})
			dot4 := view.mapObject(Coordinates{
				X: obj.Position.X - obj.Size*2,
				Y: obj.Position.Y + obj.Size,
				Z: obj.Position.Z,
			})
			view.state.screen.printLine(dot1, dot2)
			view.state.screen.printLine(dot2, dot3)
			view.state.screen.printLine(dot3, dot4)
			view.state.screen.printLine(dot4, dot1)
		}
	}
	//fmt.Printf("\n")
}

func (view *View) Loop(ctx context.Context, cancel func()) error {
	defer cancel()
	//fmt.Println("==>Loop")

	tick := time.NewTicker(time.Millisecond * time.Duration(2)).C
	count := 0
	for {
		select {
		case <-ctx.Done():
			return nil
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
		}
	}
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

type Obj_type int

const (
	Obj_Dot Obj_type = iota
	Obj_Line
	Obj_Box
	Obj_Char
	Obj_Star
)

type Object struct {
	Position Coordinates //位置
	//Direction Direction   //方向
	Type Obj_type
	Size int
}

//func (obj *Object) getPosition

type Space struct {
	Objects []Object
}

func (spc *Space) addObj(obj Object) {
	spc.Objects = append(spc.Objects, obj)
}

func NewSpace() *Space {
	//fmt.Printf("NewSpace Start")
	spc := &Space{}
	/*
		min := 0
		max := 300
		intervalX := 1
		intervalY := 12
		intervalZ := 12
		//min = min + interval
		for x := min; x <= max; x += intervalX {
			for y := min; y <= max; y += intervalY {
				for z := min; z <= max; z += intervalZ {
					obj := Object{
						Position:  Coordinates{X: x, Y: y, Z: z},
						Direction: Direction{theta: 0, phi: 0},
						Type:      Obj_Star,
						Size:      1,
					}
					spc.addObj(obj)
				}
			}
		}
	*/
	count := 1000
	w, h := termbox.Size()
	max := int((w + h) * 10)
	min := -max
	depth := (w + h) * 100
	for i := 0; i < count; i++ {
		spc.addObj(Object{
			Position: Coordinates{
				X: (min + rand.Intn(max-min)) * 2,
				Y: min + rand.Intn(max-min),
				Z: rand.Intn(depth),
			},
			Type: Obj_Dot,
			Size: 1,
		})
	}

	count = 100
	for i := 0; i < count; i++ {
		spc.addObj(Object{
			Position: Coordinates{
				X: (min + rand.Intn(max-min)) * 2,
				Y: min + rand.Intn(max-min),
				Z: rand.Intn(depth),
			},
			Type: Obj_Box,
			Size: rand.Intn(max),
		})
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
	}
}

func (state *Olion) Run(ctx context.Context) (err error) {

	go NewView(state).Loop(ctx, state.cancelFunc)
	time.Sleep(5 * time.Second)

	return nil
}
