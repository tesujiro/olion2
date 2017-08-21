package olion

import (
	"context"
	"fmt"
	"io"
	"math"
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

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

/*
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
*/

func NewScreen() *Screen {
	//w, h := getWinsize()
	w, h := termbox.Size()
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

	tick := time.NewTicker(time.Millisecond * time.Duration(100)).C
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick:
			view.eraseObjects()
			view.drawObjects()
			view.state.direction = Direction{
				theta: view.state.direction.theta + 0.01,
				phi:   view.state.direction.phi + 0.01,
			}
		}
	}
}

func (sc *Screen) printDot(dot Dot) {
	//fmt.Printf("\x1b[%v;%vH%s", sc.Height-dot.Y+1, dot.X, "X")
	termbox.SetCell(dot.X, sc.Height-dot.Y+1, 'X', termbox.ColorDefault, 1)
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
	sinPhi := math.Sin(phi)
	cosPhi := math.Cos(phi)
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
	if 0 <= dot.X && dot.X <= view.state.screen.Width && 0 <= dot.Y && dot.Y <= view.state.screen.Height {
		/*
			fmt.Printf("mapObject ObjectPosition:%v Screen:%v Position:%v Direction:%v", objPosition, myScreen, myPosition, myDirection)
			fmt.Printf(" sinTheta=%v cosTheta=%v sinPhi=%v cosPhi=%v ", sinTheta, cosTheta, sinPhi, cosPhi)
			fmt.Printf(" diffX:%v diffY:%v diffZ:%v X:%v Y:%v Z:%v", diffX, diffY, diffZ, myCoordinates.X, myCoordinates.Y, myCoordinates.Z)
			fmt.Printf(" map=>%v \n", dot)
		*/
		return &dot
	}
	return nil
}

/*
 */

func (view *View) drawObjects() {
	//fmt.Println("==>drawObjects")
	for _, obj := range view.state.space.Objects {
		//dot := Dot{X: obj.Position.X, Y: obj.Position.Y}
		if dot := view.mapObject(obj.Position); dot != nil {
			view.state.screen.printDot(*dot)
			//view.drawn = append(view.drawn, *dot)
		}
	}
	termbox.Flush()
}

func (view *View) eraseObjects() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
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
		max := 100
		intervalX := 1
		intervalY := 10
		intervalZ := 10
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
	min := 0
	max := 100
	for i := 0; i <= count; i++ {
		spc.addObj(Object{
			Position: Coordinates{
				X: min + rand.Intn(max-min),
				Y: min + rand.Intn(max-min),
				Z: min + rand.Intn(max-min),
			},
			Type: Obj_Star,
			Size: 1,
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
		position: Coordinates{X: 0, Y: 0, Z: 0},
		//direction: Direction{theta: 0, phi: 0},
		direction: Direction{theta: 10, phi: 20},
	}
}

func (state *Olion) Run(ctx context.Context) (err error) {
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	go NewView(state).Loop(ctx, state.cancelFunc)
	time.Sleep(10 * time.Second)

	return nil
}
