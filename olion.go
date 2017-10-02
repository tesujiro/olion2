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
	Objects []Exister
	Min     Coordinates
	Max     Coordinates
}

func (spc *Space) addObj(obj Exister) {
	spc.Objects = append(spc.Objects, obj)
}

func (spc *Space) randomSpace() Coordinates {
	if spc.Max.Z-spc.Min.Z > 0 {
		return Coordinates{
			X: (spc.Min.X + rand.Intn(spc.Max.X-spc.Min.X)),
			Y: (spc.Min.Y + rand.Intn(spc.Max.Y-spc.Min.Y)),
			Z: (spc.Min.Z + rand.Intn(spc.Max.Z-spc.Min.Z)),
		}
	} else {
		return Coordinates{
			X: (spc.Min.X + rand.Intn(spc.Max.X-spc.Min.X)),
			Y: (spc.Min.Y + rand.Intn(spc.Max.Y-spc.Min.Y)),
			Z: 0,
		}
	}
}

func (spc *Space) inTheSpace(c Coordinates) bool {
	return c.X >= spc.Min.X && c.X <= spc.Max.X && c.Y >= spc.Min.Y && c.Y <= spc.Max.Y && c.Z >= spc.Min.Z && c.Z <= spc.Max.Z
}

func (spc *Space) genObject(now time.Time) Exister {
	//num := rand.Intn(100)
	switch {
	//case num < 10:
	default:
		//Add SpaceShip
		return newSpaceShip(now, 500, spc.randomSpace())
		//return newStar(now, 1, spc.randomSpace())
	}
}

func (spc *Space) genBackgroundObject(now time.Time) Exister {
	//num := rand.Intn(100)
	switch {
	default:
		//Add Star
		return newStar(now, 1, spc.randomSpace())
	}
}

func NewSpace(ctx context.Context, cancel func()) *Space {
	spc := &Space{}

	w, h := termbox.Size()
	max := int((w + h) * 10)
	min := -max
	depth := (w + h) * 50

	spc.Min = Coordinates{
		X: min,
		Y: min,
		Z: 0,
	}
	spc.Max = Coordinates{
		X: max,
		Y: max,
		Z: depth,
	}
	now := time.Now()
	for i := 0; i < 50; i++ {
		obj := spc.genObject(now)
		spc.addObj(obj)
		go obj.run(ctx, cancel)
	}

	fmt.Printf("==> %v Objects\n", len(spc.Objects))
	return spc
}

func NewOuterSpace(ctx context.Context, cancel func()) *Space {
	spc := &Space{}

	w, h := termbox.Size()
	max := int((w + h) * 2)
	min := 0
	//depth := (w + h) * 100

	spc.Min = Coordinates{
		X: min,
		Y: min,
		Z: 0,
	}
	spc.Max = Coordinates{
		X: max,
		Y: max,
		Z: 0,
	}
	now := time.Now()
	for i := 0; i < 200; i++ {
		obj := spc.genBackgroundObject(now)
		spc.addObj(obj)
		go obj.run(ctx, cancel)
	}

	fmt.Printf("OuterSpace ==> %v Objects\n", len(spc.Objects))
	return spc
}

func (spc *Space) move(t time.Time, dp Coordinates) {
	downMsg := downMessage{
		time:          t,
		deltaPosition: dp,
	}
	//fmt.Printf("len(spc.Objects)=%d                                              \n", len(spc.Objects))
	for _, obj := range spc.Objects {
		//fmt.Printf("Object=%v downMsg=%v\n", obj, downMsg)
		ch := obj.downCh()
		//fmt.Printf("send message=%v type(obj.downCh())=%v\n", downMsg, reflect.TypeOf(ch))
		//obj.downCh() <- downMsg
		ch <- downMsg
		//fmt.Printf("finished send message\n")
	}
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
	readyCh    chan struct{}
	screen     *Screen
	space      *Space
	outerSpace *Space

	position Coordinates
	speed    int

	// cancelFunc is called for Exit()
	cancelFunc func()
	// Errors are stored here
	err error
}

func New(ctx context.Context, cancel func()) *Olion {
	rand.Seed(time.Now().UnixNano())

	return &Olion{
		Argv:   os.Args,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		//currentLineBuffer: NewMemoryBuffer(), // XXX revisit this
		readyCh:    make(chan struct{}),
		screen:     NewScreen(),
		space:      NewSpace(ctx, cancel),
		outerSpace: NewOuterSpace(ctx, cancel),
		//maxScanBufferSize: bufio.MaxScanTokenSize,
		position: Coordinates{X: 0, Y: 0, Z: 0},
		speed:    1,
		//cancelFunc: func() {},
	}
}

func (state *Olion) Loop(view *View, ctx context.Context, cancel func()) error {
	defer cancel()

	TermBoxChan := state.screen.TermBoxChan()
	tick := time.NewTicker(time.Millisecond * time.Duration(5)).C
	count := 0
	moveX, moveY := 0, 0
mainloop:
	for {
		//fmt.Printf("[count=%v]\n", count)
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			state.screen.clear()
			state.space.move(time.Now(), Coordinates{
				X: moveX,
				Y: moveY,
				Z: state.speed,
			})
			var c Coordinates
			if count%20 == 0 {
				c = Coordinates{X: moveX, Y: moveY, Z: 0}
			} else {
				c = Coordinates{X: 0, Y: 0, Z: 0}
			}
			state.outerSpace.move(time.Now(), c)
			view.drawBackgroundObjects()
			view.drawObjects()
			count++
			drawLine(0, 0, fmt.Sprintf("counter=%v position=%v move=(%v,%v)", count, state.position, moveX, moveY))
			state.screen.flush()
		case ev := <-TermBoxChan:
			if ev.Type == termbox.EventKey {
				switch ev.Key {
				case termbox.KeyEsc:
					break mainloop // Esc で実行終了
				case termbox.KeyArrowUp:
					moveY--
				case termbox.KeyArrowDown:
					moveY++
				case termbox.KeyArrowLeft:
					moveX++
				case termbox.KeyArrowRight:
					moveX--
					//dafault:
				}
			}
		}
	}
	return nil
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
	go state.Loop(NewView(state), ctx, cancel)
	//time.Sleep(5 * time.Second)

	// Alright, done everything we need to do automatically. We'll let
	// the user play with peco, and when we receive notification to
	// bail out, the context should be canceled appropriately
	<-ctx.Done()

	return nil
}
