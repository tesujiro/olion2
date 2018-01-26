package olion

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

func (state *Olion) move(spc *Space, t time.Time, dp Coordinates, ctx context.Context, cancel func()) []upMessage {
	downMsg := downMessage{
		time:          t,
		deltaPosition: dp,
	}
	upMsgs := []upMessage{}
	now := time.Now()

	// send downMsg , and get flying objects and bombs
	flyings := []Exister{}
	bombs := []Exister{}
	for _, obj := range spc.Objects {
		ch := obj.downCh()
		ch <- downMsg
		if !obj.isBomb() {
			flyings = append(flyings, obj)
		} else {
			bombs = append(bombs, obj)
		}
	}

	// get msg from flying objects
	for _, flying := range flyings {
		upMsg := <-flying.upCh()
		upMsgs = append(upMsgs, upMsg)
		// Throw a bomb from enemy.
		if !flying.isExploding() && flying.hasBomb() && state.screen.distance(flying.getPosition(), Coordinates{}) < flying.getThrowBombDistance() {
			sp1 := state.speed
			position := flying.getPosition()
			distance := state.screen.distance(position, Coordinates{})
			debug.Printf("Enemy Bomb!! self=%v speed=%v position=%v distance=%v\n", &flying, sp1, position, distance)
			k := 0
			if position.Z != 0 {
				//k = distance * 80 * 1000 / position.Z
				k = distance * 80 / position.Z // 80 : z axis speed of enemy bombs
			}
			//speed := Coordinates{X: -position.X*k/distance/1000 + sp1.X, Y: -position.Y*k/distance/1000 + sp1.Y, Z: -position.Z*k/distance/1000 + sp1.Z}
			speed := position.ScaleBy(-k).Div(distance).Add(sp1)
			//debug.Printf("speed=%v\n", speed)
			newObj := newEnemyBomb(now, 1000, position, speed)
			newObj.setBomber(flying)
			state.space.addObj(newObj)
			flying.removeBomb()
			go newObj.run(ctx, cancel)
		}
		// Todo:敵同士の攻撃
	}

	// Stop Vibration
	if state.exploding {
		deltaTime := float64(time.Now().Sub(state.explodedAt) / time.Millisecond)
		if deltaTime > float64(3e3) {
			// Stop vibration 3 sec. after explosion.
			state.screen.Vibration = 0
			state.exploding = false
		}
	}

	between := func(a, b, c int) bool {
		return (a < b && b <= c) || (a > b && b >= c)
	}
	min := func(a, b int) int {
		if a >= b {
			return b
		} else {
			return a
		}
	}
	max := func(a, b int) int {
		if a >= b {
			return a
		} else {
			return b
		}
	}
	cross := func(a Exister, b Coordinates) bool {
		aSize := a.getSize() / 2
		aAt := a.getPosition()
		aPrevAt := a.getPrevPosition()
		return between(min(aAt.X, aPrevAt.X)-aSize, b.X, max(aAt.X, aPrevAt.X)+aSize) && between(min(aAt.Y, aPrevAt.Y)-aSize, b.Y, max(aAt.Y, aPrevAt.Y)+aSize) && between(aPrevAt.Z, b.Z, aAt.Z)
	}

	// receive msg from bombs and judge explosion
	for _, bomb := range bombs {
		upMsg := <-bomb.upCh()
		upMsgs = append(upMsgs, upMsg)
		if cross(bomb, Coordinates{}) && state.screen.distance(Coordinates{}, bomb.getPosition()) <= bomb.getSize() {
			debug.Printf("my object exploded!!!\n")
			state.score--
			state.screen.Vibration = 3
			state.explodedAt = time.Now()
			state.exploding = true
		}

	L:
		for _, flying := range flyings {
			if bomb.getBomber() != flying && cross(bomb, flying.getPosition()) && state.screen.distance(flying.getPosition(), bomb.getPosition()) <= bomb.getSize() {
				debug.Printf("the flying object exploded!!!\n")
				debug.Printf("bomb@%v flying@%v distance=%v\n", bomb.getPosition(), flying.getPosition(), state.screen.distance(flying.getPosition(), bomb.getPosition()))
				state.score++
				flying.explode()
				spc.deleteObj(bomb)
				break L
			}
		}
	}

	for _, obj := range spc.Objects {
		// stop flying object explosion
		if obj.isExploding() {
			deltaTime := float64(time.Now().Sub(obj.getExplodedTime()) / time.Millisecond)
			if deltaTime > float64(1e4) {
				// Delete 10 sec. after explosion.
				spc.deleteObj(obj)
				newObj := spc.GenFunc(now)
				spc.addObj(newObj)
				go newObj.run(ctx, cancel)
			}
		}
		// if objct is out of the Space , remove it and create new one
		if !spc.inTheSpace(obj.getPosition()) {
			//if fmt.Sprintf("%v", reflect.TypeOf(obj)) != "*olion.Star" {
			//debug.Printf("objct(%v) is out of the Space (%v), remove and create new one\n", reflect.TypeOf(obj), obj.getPosition())
			//}
			spc.deleteObj(obj)
			if !obj.isBomb() {
				//debug.Printf("objct is not a bomb\n")
				newObj := spc.GenFunc(now)
				spc.addObj(newObj)
				go newObj.run(ctx, cancel)
				newObj.downCh() <- downMsg
				upMsg := <-newObj.upCh()
				upMsgs = append(upMsgs, upMsg)
			}
		}
	}

	return upMsgs
}

type Olion struct {
	Argv        []string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	Debug       bool
	Palette     bool
	Pause       bool
	debugWindow *Window
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
	mobile
	maxBomb     int
	curBomb     int
	score       int
	dispFps     int
	dispFpsUnix int64
	curFps      int
	curFpsUnix  int64
	//vibration int
	explodedAt time.Time
	exploding  bool

	// cancelFunc is called for Exit()
	cancelFunc func()
	// Errors are stored here
	err error
}

func New(ctx context.Context, cancel func()) *Olion {
	rand.Seed(time.Now().UnixNano())
	debug := flag.Bool("d", false, "Debug Mode")
	palette := flag.Bool("p", false, "Color Palette Mode")
	flag.Parse()
	screen := NewScreen()
	newDebugWriter(ctx)
	InitColor()

	return &Olion{
		Argv:        os.Args,
		Stderr:      os.Stderr,
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Debug:       *debug,
		Palette:     *palette,
		Pause:       false,
		debugWindow: newDebugWindow(screen),
		//currentLineBuffer: NewMemoryBuffer(), // XXX revisit this
		readyCh:    make(chan struct{}),
		screen:     screen,
		space:      NewSpace(ctx, cancel),
		outerSpace: NewOuterSpace(ctx, cancel),
		//maxScanBufferSize: bufio.MaxScanTokenSize,
		position: Coordinates{X: 0, Y: 0, Z: 0},
		mobile:   mobile{speed: Coordinates{X: 0, Y: 0, Z: 20}, time: time.Now()},
		maxBomb:  4,
		curBomb:  0,
		score:    0,
		//vibration: 0,
		//cancelFunc: func() {},
	}
}

func (state *Olion) drawConsole(count int) {
	unix := time.Now().Unix()
	if unix == state.curFpsUnix {
		state.curFps++
	} else {
		state.dispFps = state.curFps
		state.dispFpsUnix = state.curFpsUnix
		state.curFps = 0
		state.curFpsUnix = unix
	}
	state.screen.printString(&Dot{0, 0}, fmt.Sprintf("%v frameRate=%vfps counter=%v move=%v bombs=%v", time.Unix(state.dispFpsUnix, 0), state.dispFps, count, state.speed, state.curBomb))

	//state.disp_number(123456789)
	start := Dot{0, state.screen.Height - 5}
	//disp_number(start, state.score)
	disp_string(start, fmt.Sprintf("SCORE:%v", state.score))
	x, y := state.screen.Width/2+1, state.screen.Height/2+1
	for i := 0; i < state.maxBomb-state.curBomb; i++ {
		state.screen.printString(&Dot{x, y}, "**")
		state.screen.printString(&Dot{x, y + 1}, "**")
		x += 3
		y += 0
	}
}

type debugWriter struct {
	//w      io.Writer
	//buff    [][]byte
	buff    []string
	curLine int

	// Goroutine Implementation
	writeChan chan string
	writeDone chan struct{}
	readReq   chan int
	readChan  chan string
	//readDone  chan struct{}
}

var debug *debugWriter

func newDebugWriter(ctx context.Context) {
	size := 1000
	d := &debugWriter{
		buff:      make([]string, size),
		curLine:   0,
		writeChan: make(chan string),
		writeDone: make(chan struct{}),
		readReq:   make(chan int),
		readChan:  make(chan string),
		//readDone:  make(chan struct{}),
	}

	go func() {
	L:
		for {
			select {
			case str := <-d.writeChan:
				lines := strings.Count(str, "\n")
				for idx, line := range strings.Split(str, "\n") {
					if idx < lines || len(line) > 0 {
						//d.buff[d.curLine] = fmt.Sprintf("[%v]:%v", idx, line) //Todo: bad performance
						d.buff[d.curLine] = line //Todo: bad performance
						d.curLine = (d.curLine + 1) % len(d.buff)
					}
				}
				d.writeDone <- struct{}{}
			case size := <-d.readReq:
				firstLine := (d.curLine + len(d.buff) - size) % len(d.buff)
				for i := 0; i < size; i++ {
					idx := (firstLine + i) % len(d.buff)
					msg := d.buff[idx]
					//msg := strconv.Itoa(idx) + ":" + msg
					d.readChan <- msg
				}
			case <-ctx.Done():
				break L
			}
		}
	}()

	debug = d
}

func (d *debugWriter) Write(p []byte) (int, error) {
	d.writeChan <- string(p)
	<-d.writeDone
	return len(p), nil
}

func (d *debugWriter) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(d, format, a...)
}

type Window struct {
	screen *Screen
	width  int
	height int
	StartX int
	StartY int
	cursor int
}

func newDebugWindow(screen *Screen) *Window {
	return &Window{
		width:  80,
		height: 50,
		screen: screen,
		StartX: 5,
		StartY: 5,
	}
}

/*
func (state *Olion) Printf(format string, a ...interface{}) (n int, err error) {
	d := state.debugWriter
	return fmt.Fprintf(d, format, a...)
}
*/

func (state *Olion) drawDebugInfo() {
	//d := state.debugWriter
	d := debug
	w := state.debugWindow

	//draw debug window frame
	for x := w.StartX - 1; x <= w.StartX+w.width; x++ {
		w.screen.printString(&Dot{x, w.StartY - 1}, "+")
		w.screen.printString(&Dot{x, w.StartY + w.height}, "+")
	}
	for y := w.StartY; y < w.StartX+w.height; y++ {
		w.screen.printString(&Dot{w.StartX - 1, y}, "+")
		w.screen.printString(&Dot{w.StartX + w.width, y}, "+")
	}

	//print debug buffer
	d.readReq <- w.height
	for i := 0; i < w.height; i++ {
		msg := <-d.readChan
		//msg = strconv.Itoa(i) + ":" + msg
		if len(msg) > w.width {
			msg = msg[:w.width]
		}
		w.screen.printString(&Dot{w.StartX, w.StartY + i}, msg)
	}
}

func (state *Olion) setStatus() {
	//count bombs
	bombs := 0
	for _, obj := range state.space.Objects {
		if obj.isBomb() {
			bombs++
		}
	}
	state.curBomb = bombs
}

func (state *Olion) Loop(view *View, ctx context.Context, cancel func()) error {
	defer cancel()

	TermBoxChan := state.screen.TermBoxChan()
	//tick := time.NewTicker(time.Millisecond * time.Duration(5)).C
	tick := time.NewTicker(time.Millisecond * time.Duration(20)).C
	count := 0
	fireBomb := false
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			//debug.Printf("tick\n")
			if state.Pause {
				continue mainloop
			}
			state.screen.clear()
			//OuterSpace
			speed := Coordinates{X: state.speed.X, Y: state.speed.Y, Z: 0}
			upMsgsOuterSpace := state.move(state.outerSpace, time.Now(), speed, ctx, cancel)
			view.draw(upMsgsOuterSpace)
			//Space
			now := time.Now()
			if fireBomb {
				debug.Printf("newBomb\n")
				speed := Coordinates{state.speed.X, state.speed.Y, state.speed.Z + 80}
				newObj := newBomb(now, 1000, Coordinates{}, speed)
				state.space.addObj(newObj)
				go newObj.run(ctx, cancel)
				fireBomb = false
			}
			forward := state.getDistance(now)
			upMsgs := state.move(state.space, now, forward, ctx, cancel)
			view.draw(upMsgs)
			count++
			state.setStatus()
			state.drawConsole(count)
			//state.screen.printLine(&Dot{X: 10, Y: 12}, &Dot{X: 20, Y: 17}, ColorRed)
			//state.screen.printLine(&Dot{X: 10, Y: 32}, &Dot{X: 15, Y: 42}, ColorRed)
			//state.screen.printPolygon([]Dot{Dot{X: 10, Y: 10}, Dot{X: 40, Y: 50}, Dot{X: 60, Y: 100}, Dot{X: 10, Y: 40}}, colors.name("White").Attribute(), true)
			//state.screen.printPolygon([]Dot{Dot{X: 10, Y: 10}, Dot{X: 40, Y: 50}, Dot{X: 60, Y: 100}, Dot{X: 10, Y: 40}}, colors.name("Black").Attribute(), false)
			/*
				var d1, d2, d3 Dot
				d1, d2, d3 = Dot{X: 10, Y: 10}, Dot{X: 15, Y: 20}, Dot{X: 20, Y: 10}
				debug.Printf("d1=%v d2=%v d3=%v\n", d1, d2, d3)
				state.screen.printTriangle([]Dot{d1, d2, d3}, ColorBlack)
				d1, d2, d3 = Dot{X: 35, Y: 20}, Dot{X: 40, Y: 10}, Dot{X: 30, Y: 10}
				debug.Printf("d1=%v d2=%v d3=%v\n", d1, d2, d3)
				state.screen.printTriangle([]Dot{d1, d2, d3}, ColorBlack)
			*/
			//debug.Printf("len(colors)=%v color.red=%v id=%v ColorRed=%v\n", len(colors), colors.name("Red"), colors.name("Red").ColorId, ColorRed)
			//debug.Printf("len(colors)=%v color.black=%v id=%v ColorBlack=%v\n", len(colors), colors.name("Black"), colors.name("Black").ColorId, ColorBlack)
			//debug.Printf("typeOf(ColorId)=%v typeOf(ColorBlack)=%v\n", reflect.TypeOf(colors.name("Black").ColorId), reflect.TypeOf(ColorBlack))
			//debug.Printf("tick ->End\n\n")
			if state.Palette == true {
				state.drawColorPalette()
			}
			if state.Debug == true {
				state.drawDebugInfo()
			}
			state.screen.flush()
		case ev := <-TermBoxChan:
			//debug.Printf("TermBoxChan\n")
			upspeed := func(speed int, delta int) int {
				limit := 80
				switch {
				case delta > 0 && speed+delta <= limit:
					return speed + delta
				case delta < 0 && speed+delta >= -limit:
					return speed + delta
				default:
					return speed
				}
			}
			if ev.Type == termbox.EventKey {
				switch ev.Key {
				case termbox.KeyEsc:
					break mainloop // Esc で実行終了
				case termbox.KeyArrowUp:
					state.speed.Y = upspeed(state.speed.Y, -10)
				case termbox.KeyArrowDown:
					state.speed.Y = upspeed(state.speed.Y, 10)
				case termbox.KeyArrowLeft:
					state.speed.X = upspeed(state.speed.X, 10)
				case termbox.KeyArrowRight:
					state.speed.X = upspeed(state.speed.X, -10)
				case termbox.KeySpace:
					state.speed.Z = upspeed(state.speed.Z, 10)
				case termbox.KeyTab, termbox.KeyCtrlSpace:
					if state.speed.Z > 0 {
						state.speed.Z -= 10
					}
					//state.speed.Z = upspeed(state.speed.Z, -10)
				case termbox.KeyEnter:
					if state.curBomb < state.maxBomb {
						state.curBomb++
						fireBomb = true
					}
				case termbox.KeyF1:
					if state.screen.Vibration == 0 {
						state.screen.Vibration = 1
					} else {
						state.screen.Vibration = 0
					}
				case termbox.KeyF2:
					if state.Pause == true {
						state.Pause = false
					} else {
						state.Pause = true
					}
				default:
				}
			}
			//debug.Printf("TermBoxChan ->End\n")
		}
	}
	return nil
}

func (state *Olion) Run(ctx context.Context) (err error) {

	//InitColor()
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

	// Alright, done everything we need to do automatically. We'll let
	// the user play with peco, and when we receive notification to
	// bail out, the context should be canceled appropriately
	<-ctx.Done()

	return nil
}
