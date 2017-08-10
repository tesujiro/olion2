package olion

import (
	"context"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type Dot struct {
	X int
	Y int
}

type Screen struct {
	Width  int
	Height int
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
	//fmt.Printf("W=%v H=%v\n", int(w), int(h))
	return &Screen{Width: int(w), Height: int(h)}
}

type View struct {
	state *Olion
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
			view.drawScreen()
		}
	}
}

func (sc *Screen) printDot(dot Dot) {
	fmt.Printf("\x1b[%v;%vH%s", sc.Height-dot.Y+1, dot.X, "X")
}

func (view *View) drawScreen() {
	//fmt.Println("==>drawScreen")
	for _, obj := range view.state.space.Objects {
		dot := Dot{X: obj.Position.X, Y: obj.Position.Y}
		view.state.screen.printDot(dot)
	}
	fmt.Printf("\n")
}

type Coordinates struct {
	X int
	Y int
	Z int
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
	Direction Coordinates //方向
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
	fmt.Printf("NewSpace Start\n")
	spc := &Space{}
	min := -1000
	max := 1000
	interval := 50
	for x := min; x <= max; x += interval {
		for y := min; y <= max; y += interval {
			for z := min; z <= max; z += interval {
				obj := Object{
					Position:  Coordinates{X: x, Y: y, Z: z},
					Direction: Coordinates{X: 0, Y: 0, Z: 0},
					Type:      Obj_Star,
					Size:      1,
				}
				spc.addObj(obj)
			}
		}
	}

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
	direction Coordinates

	// cancelFunc is called for Exit()
	cancelFunc func()
	// Errors are stored here
	err error
}

func New() *Olion {
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
		direction: Coordinates{X: 0, Y: 0, Z: 1},
	}
}

func (state *Olion) Run(ctx context.Context) (err error) {
	go NewView(state).Loop(ctx, state.cancelFunc)
	time.Sleep(3 * time.Second)

	return nil
}
