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
	GenFunc func(time.Time) Exister
}

func (spc *Space) addObj(obj Exister) {
	spc.Objects = append(spc.Objects, obj)
}

func (spc *Space) deleteObj(obj Exister) {
	objects := []Exister{}
	for _, v := range spc.Objects {
		if v != obj {
			objects = append(objects, v)
		}
	}
	spc.Objects = objects
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
	spc.GenFunc = spc.genObject

	w, h := termbox.Size()
	max := int((w + h) * 30)
	min := -max
	depth := (w + h) * 40

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
	for i := 0; i < 20; i++ {
		obj := spc.GenFunc(now)
		spc.addObj(obj)
		go obj.run(ctx, cancel)
	}

	fmt.Printf("==> %v Objects\n", len(spc.Objects))
	return spc
}

func NewOuterSpace(ctx context.Context, cancel func()) *Space {
	spc := &Space{}
	spc.GenFunc = spc.genBackgroundObject

	w, h := termbox.Size()
	max := int((w + h) * 20)
	min := -max
	depth := max

	spc.Min = Coordinates{
		X: min,
		Y: min,
		Z: 0,
	}
	spc.Max = Coordinates{
		X: max,
		Y: max,
		Z: depth / 20,
	}
	now := time.Now()
	for i := 0; i < 10; i++ {
		obj := spc.GenFunc(now)
		spc.addObj(obj)
		go obj.run(ctx, cancel)
	}

	fmt.Printf("OuterSpace ==> %v Objects\n", len(spc.Objects))
	return spc
}

func (spc *Space) move(t time.Time, dp Coordinates, ctx context.Context, cancel func()) []upMessage {
	downMsg := downMessage{
		time:          t,
		deltaPosition: dp,
	}
	upMsgs := []upMessage{}
	//fmt.Printf("len(spc.Objects)=%d                                              \n", len(spc.Objects))
	for _, obj := range spc.Objects {
		//fmt.Printf("Object=%v downMsg=%v\n", obj, downMsg)
		ch := obj.downCh()
		//fmt.Printf("send message=%v type(obj.downCh())=%v\n", downMsg, reflect.TypeOf(ch))
		//obj.downCh() <- downMsg
		ch <- downMsg
		//fmt.Printf("finished send message\n")
	}
	now := time.Now()
	for _, obj := range spc.Objects {
		//fmt.Printf("upMsg := <-obj.upCh()")
		upMsg := <-obj.upCh()
		//fmt.Printf("(upMsg:%v)\n", upMsg)
		//if objct is out of the Space , remove and create new one
		if !spc.inTheSpace(upMsg.position) {
			spc.deleteObj(obj)
			newObj := spc.GenFunc(now)
			spc.addObj(newObj)
			go newObj.run(ctx, cancel)
			newObj.downCh() <- downMsg
			upMsg = <-newObj.upCh()
		}
		upMsgs = append(upMsgs, upMsg)
	}
	return upMsgs
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
	//speed    int
	speed Coordinates
	time  time.Time

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
		//speed:    1,
		speed: Coordinates{X: 0, Y: 0, Z: 1},
		time:  time.Now(),
		//cancelFunc: func() {},
	}
}

func (state *Olion) getTime() time.Time {
	return state.time
}

func (state *Olion) setTime(t time.Time) {
	state.time = t
}

func (state *Olion) getSpeed() Coordinates {
	return state.speed
}

func (state *Olion) Loop(view *View, ctx context.Context, cancel func()) error {
	defer cancel()

	TermBoxChan := state.screen.TermBoxChan()
	tick := time.NewTicker(time.Millisecond * time.Duration(5)).C
	count := 0
	//moveX, moveY := 0, 0
mainloop:
	for {
		//fmt.Printf("[count=%v]\n", count)
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			state.screen.clear()
			upMsgs := state.space.move(time.Now(), Coordinates{
				X: state.speed.X,
				Y: state.speed.Y,
				Z: state.speed.Z,
			}, ctx, cancel)
			var c Coordinates
			if count%20 == 0 {
				//c = Coordinates{X: moveX, Y: moveY, Z: 0}
				c = Coordinates{X: state.speed.X, Y: state.speed.Y, Z: 0}
			} else {
				c = Coordinates{X: 0, Y: 0, Z: 0}
			}
			upMsgsOuterSpace := state.outerSpace.move(time.Now(), c, ctx, cancel)
			//view.draw(state.outerSpace.Objects)
			view.draw(upMsgsOuterSpace)
			//view.draw(state.space.Objects)
			view.draw(upMsgs)
			count++
			drawLine(0, 0, fmt.Sprintf("counter=%v position=%v move=%v", count, state.position, state.speed))
			state.screen.flush()
		case ev := <-TermBoxChan:
			if ev.Type == termbox.EventKey {
				switch ev.Key {
				case termbox.KeyEsc:
					break mainloop // Esc で実行終了
				case termbox.KeyArrowUp:
					//moveY--
					state.speed.Y--
				case termbox.KeyArrowDown:
					//moveY++
					state.speed.Y++
				case termbox.KeyArrowLeft:
					//moveX++
					state.speed.X++
				case termbox.KeyArrowRight:
					//moveX--
					state.speed.X--
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
