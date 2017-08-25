package olion

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"syscall"
	"time"
	"unsafe"

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

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getWinsize() (uint, uint) {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}

	//fmt.Printf("Xpixel=%v Ypixel=%v\n", ws.Xpixel, ws.Ypixel)
	return uint(ws.Col), uint(ws.Row)
}

func NewScreen() *Screen {
	w, h := getWinsize()
	//w, h := termbox.Size()    //????? no value
	d := 10
	fmt.Printf("W=%v H=%v\n", int(w), int(h))
	return &Screen{Width: int(w), Height: int(h), Distance: d}
}

type View struct {
	state *Olion
	//drawn []Dot
}

func NewView(state *Olion) *View {
	return &View{state: state}
}

func (view *View) Loop(ctx context.Context, cancel func()) error {
	defer cancel()
	//fmt.Println("==>Loop")

	tick := time.NewTicker(time.Millisecond * time.Duration(1)).C
	count := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick:
			//view.eraseObjects()
			view.drawObjects()
			view.state.direction = Direction{
				theta: view.state.direction.theta + 0.01,
				phi:   view.state.direction.phi + 0.01,
			}
			/*
				view.state.position = Coordinates{
					X: view.state.position.X + 1,
				}
			*/
			count++
			drawLine(0, 0, fmt.Sprintf("counter=%v", count))
			termbox.Flush()
		}
	}
}

func (sc *Screen) cover(dot Dot) bool {
	return 0 <= dot.X && dot.X <= sc.Width && 0 <= dot.Y && dot.Y <= sc.Height
}

func (sc *Screen) printDot(dot Dot) {
	//fmt.Printf("\x1b[%v;%vH%s", sc.Height-dot.Y+1, dot.X, "X")
	if sc.cover(dot) {
		termbox.SetCell(dot.X, sc.Height-dot.Y+1, ' ', termbox.ColorDefault, 1)
	}
	//drawLine(10, 10, fmt.Sprintf("counter="))
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

func drawLine(x, y int, str string) {
	color := termbox.ColorDefault
	//backgroundColor := termbox.ColorDefault
	runes := []rune(str)

	for i := 0; i < len(runes); i += 1 {
		termbox.SetCell(x+i, y, runes[i], color, 1)
	}
}

/*
func (sc *Screen) eraseDot(dot Dot) {
	fmt.Printf("\x1b[%v;%vH%s", sc.Height-dot.Y+1, dot.X, " ")
}
*/

/*
func (view *View) mapObject(objPosition Coordinates) *Dot {
	myScreen := view.state.screen
	myPosition := view.state.position
	myDirection := view.state.diretion
	fmt.Printf("mapObject ObjectPosition:%v Screen:%v Position:%v Direction:%v", objPosition, myScreen, myPosition, myDirection)
	// reference http://www.geocities.co.jp/SiliconValley-Bay/4543/Rubic/Mathematics/Mathematics-5_1.html

	return nil
}
*/

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

func (view *View) mapObject(objPosition Coordinates) *Dot {
	myScreen := view.state.screen
	myPosition := view.state.position
	myDirection := view.state.direction
	//fmt.Printf("mapObject ObjectPosition:%v Screen:%v Position:%v Direction:%v", objPosition, view.state.screen, view.state.position, view.state.direction)
	// reference http://www.geocities.co.jp/SiliconValley-Bay/4543/Rubic/Mathematics/Mathematics-5_1.html
	/*
		a := math.Sqrt(float64(math.Pow(float64(objPosition.X-myPosition.X), float64(2)) + math.Pow(float64(objPosition.Y-myPosition.Y), float64(2))))
		b := math.Sqrt(float64(math.Pow(a, float64(2)) + math.Pow(float64(objPosition.Z-myPosition.Z), float64(2))))
		sinTheta := float64(-objPosition.Y+myPosition.Y) / a
		cosTheta := float64(-objPosition.X+myPosition.X) / a
		sinPhi := float64(-objPosition.Z+myPosition.Z) / b
		cosPhi := a / b
	*/
	theta := float64(myDirection.theta) / float64(180) * math.Pi
	phi := float64(myDirection.phi) / float64(180) * math.Pi
	sinTheta := math.Sin(theta)
	cosTheta := math.Cos(theta)
	//sinTheta := sin(theta)
	//cosTheta := cos(theta)
	sinPhi := math.Sin(phi)
	cosPhi := math.Cos(phi)
	//sinPhi := sin(phi)
	//cosPhi := cos(phi)
	diffX := float64(objPosition.X - myPosition.X)
	diffY := float64(objPosition.Y - myPosition.Y)
	diffZ := float64(objPosition.Z - myPosition.Z)
	//diffX := float64(-myPosition.X)
	//diffY := float64(-myPosition.Y)
	//diffZ := float64(-myPosition.Z)
	myCoordinates := Coordinates{
		X: int(diffX*(-sinTheta) + diffY*cosTheta),
		Y: int(diffX*(-sinPhi*cosTheta) + diffY*(-sinPhi*sinTheta) + diffZ*cosPhi),
		Z: int(diffX*(-cosPhi*cosTheta) + diffY*(-cosPhi*sinTheta) + diffZ*(-sinPhi)),
	}
	if myCoordinates.Z == 0 {
		return nil
	}
	dot := Dot{
		//X: int(myCoordinates.X / myCoordinates.Z * view.state.screen.Width),
		//Y: int(myCoordinates.Y / myCoordinates.Z * view.state.screen.Height),
		X: int(myCoordinates.X * myCoordinates.Z / myScreen.Distance),
		Y: int(myCoordinates.Y * myCoordinates.Z / myScreen.Distance),
	}
	/*
		if 0 <= dot.X && dot.X <= view.state.screen.Width && 0 <= dot.Y && dot.Y <= view.state.screen.Height {
	*/
	//fmt.Printf("dot=%v\n", dot)
	/*
		fmt.Printf("mapObject ObjectPosition:%v Screen:%v Position:%v Direction:%v", objPosition, myScreen, myPosition, myDirection)
		fmt.Printf(" sinTheta=%v cosTheta=%v sinPhi=%v cosPhi=%v ", sinTheta, cosTheta, sinPhi, cosPhi)
		fmt.Printf(" diffX:%v diffY:%v diffZ:%v X:%v Y:%v Z:%v", diffX, diffY, diffZ, myCoordinates.X, myCoordinates.Y, myCoordinates.Z)
		fmt.Printf(" map=>%v \n", dot)
	*/
	return &dot
	/*
		}
		return nil
	*/
}

/*
 */

func (view *View) drawObjects() {
	//fmt.Printf("\n==>drawObjects(%v)\n", len(view.state.space.Objects))

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
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
				X: obj.Position.X + obj.Size,
				Y: obj.Position.Y + obj.Size,
				Z: obj.Position.Z,
			})
			dot2 := view.mapObject(Coordinates{
				X: obj.Position.X + obj.Size,
				Y: obj.Position.Y - obj.Size,
				Z: obj.Position.Z,
			})
			dot3 := view.mapObject(Coordinates{
				X: obj.Position.X - obj.Size,
				Y: obj.Position.Y - obj.Size,
				Z: obj.Position.Z,
			})
			dot4 := view.mapObject(Coordinates{
				X: obj.Position.X - obj.Size,
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

func (view *View) eraseObjects() {
	/*
		for _, dot := range view.drawn {
			view.state.screen.eraseDot(dot)
		}
		view.drawn = nil
	*/
}

type Coordinates struct {
	X int
	Y int
	Z int
}

type Direction struct {
	theta float64
	phi   float64
}

type Obj_type int

const (
	Obj_Dot Obj_type = iota
	Obj_Line
	Obj_Box
	Obj_Char
	Obj_Star
)

type Object struct {
	Position  Coordinates //位置
	Direction Direction   //方向
	Type      Obj_type
	Size      int
}

//func (obj *Object) getPosition

type Space struct {
	Objects []Object
}

func (spc *Space) addObj(obj Object) {
	spc.Objects = append(spc.Objects, obj)
}

func NewSpace() *Space {
	fmt.Printf("NewSpace Start")
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
	w, h := getWinsize()
	min := 0
	max := int((w + h) / 3)
	for i := 0; i < count; i++ {
		spc.addObj(Object{
			Position: Coordinates{
				X: min + rand.Intn(max-min),
				Y: min + rand.Intn(max-min),
				Z: min + rand.Intn(max-min),
			},
			Type: Obj_Dot,
			Size: 1,
		})
	}

	spc.addObj(Object{
		Position: Coordinates{
			X: 50,
			Y: 50,
			Z: 50,
		},
		Type: Obj_Box,
		Size: 20,
	})

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

	position  Coordinates
	direction Direction

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
		position:  Coordinates{X: 0, Y: 0, Z: 0},
		direction: Direction{theta: 0, phi: 0},
		//direction: Direction{theta: 10, phi: 20},
	}
}

func (state *Olion) Run(ctx context.Context) (err error) {

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	go NewView(state).Loop(ctx, state.cancelFunc)
	time.Sleep(6 * time.Second)

	return nil
}
