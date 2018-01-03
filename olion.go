package olion

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strings"
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
	num := rand.Intn(100)
	switch {
	case true:
		return newSpaceShip(now, 500, spc.randomSpace())
	//return newFramedRectangle(now, 1000, spc.randomSpace())
	case num < 20:
		return newBox(now, 500, spc.randomSpace())
	case num < 40:
		return newBox2(now, 800, spc.randomSpace())
	case num < 60:
		return newBox3(now, 800, spc.randomSpace())
	default:
		//Add SpaceShip
		return newSpaceShip(now, 500, spc.randomSpace())
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
	//for i := 0; i < 10; i++ {
	for i := 0; i < 3; i++ {
		obj := spc.GenFunc(now)
		spc.addObj(obj)
		go obj.run(ctx, cancel)
	}

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
		//Z: depth / 20,
		Z: depth / 10,
	}
	now := time.Now()
	for i := 0; i < 10; i++ {
		obj := spc.GenFunc(now)
		spc.addObj(obj)
		go obj.run(ctx, cancel)
	}

	//fmt.Printf("OuterSpace ==> %v Objects\n", len(spc.Objects))
	return spc
}

func (state *Olion) move(spc *Space, t time.Time, dp Coordinates, ctx context.Context, cancel func()) []upMessage {
	downMsg := downMessage{
		time:          t,
		deltaPosition: dp,
	}
	upMsgs := []upMessage{}
	now := time.Now()
	flyings := []Exister{}
	bombs := []Exister{}

	for _, obj := range spc.Objects {
		ch := obj.downCh()
		ch <- downMsg
		if !obj.isBomb() {
			if obj.isExploding() {
				deltaTime := float64(time.Now().Sub(obj.getExplodedTime()) / time.Millisecond)
				if deltaTime > float64(1e4) {
					// Delete 10 sec. after explosion.
					debug.Printf("Delete 10 sec. after explosion.\n")
					spc.deleteObj(obj)
				}
			}
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
		if flying.hasBomb() && state.screen.distance(flying.getPosition(), Coordinates{}) < flying.getThrowBombDistance() {
			debug.Printf("Enemy Bomb!!\n")
			sp1 := state.speed
			sp2 := flying.getSpeed()
			debug.Printf("self speed=%v\n", sp1)
			debug.Printf("enemy speed=%v\n", sp2)
			position := flying.getPosition()
			distance := state.screen.distance(position, Coordinates{})
			debug.Printf("position=%v distance=%v\n", position, distance)
			k := 0
			if position.Z != 0 {
				k = distance * 80 * 1000 / position.Z
			}
			//speed := Coordinates{X: -position.X*k/distance/1000 - sp1.X, Y: -position.Y*k/distance/1000 - sp1.Y, Z: -position.Z*k/distance/1000 - sp1.Z} //BUG
			speed := Coordinates{X: -position.X*k/distance/1000 + sp1.X, Y: -position.Y*k/distance/1000 + sp1.Y, Z: -position.Z*k/distance/1000 + sp1.Z}
			debug.Printf("speed=%v\n", speed)
			newObj := newEnemyBomb(now, 2000, position, speed)
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
			// Stop 3 sec. after explosion.
			state.screen.Vibration = 0
			state.exploding = false
		}
	}

	// get msg from bombs and judge explosion
	for _, bomb := range bombs {
		upMsg := <-bomb.upCh()
		upMsgs = append(upMsgs, upMsg)
		bombAt := bomb.getPosition()
		debug.Printf("Bomb At %v Distance %v\n", bombAt, state.screen.distance(bombAt, Coordinates{}))
		bombPrevAt := bomb.getPrevPosition()
		between := func(a, b, c int) bool {
			return (a <= b && b <= c) || (a >= b && b >= c)
		}
		// Judge Explosion of Bombs and the first view object
		if between(bombPrevAt.Z, 0, bombAt.Z) && state.screen.distance(Coordinates{}, bomb.getPosition()) <= bomb.getSize() {
			debug.Printf("self object exploded!!!\n")
			state.screen.Vibration = 3
			state.explodedAt = time.Now()
			state.exploding = true
		}

	L:
		//debug.Printf("flying.getPosition()=%v bomb.getPosition()=%v bomb.getPrevPosition=%v\n", flying.getPosition(), bomb.getPosition(), bomb.getPrevPosition())
		// Judge Explosion of Bombs and Flying Objects
		for _, flying := range flyings {
			flyingAt := flying.getPosition()
			if between(bombPrevAt.Z, flyingAt.Z, bombAt.Z) && state.screen.distance(flying.getPosition(), bomb.getPosition()) <= bomb.getSize() {
				debug.Printf("the flying object exploded!!!\n")
				//fmt.Printf("distance=%v size=%v\n", distance(flying.getPosition(), bomb.getPosition()), bomb.getSize())
				state.score++
				flying.explode()
				spc.deleteObj(bomb)
				break L
			}
		}
	}

	// if objct is out of the Space , remove and create new one
	for _, obj := range spc.Objects {
		if !spc.inTheSpace(obj.getPosition()) {
			if fmt.Sprintf("%v", reflect.TypeOf(obj)) != "*olion.Star" {
				debug.Printf("objct(%v) is out of the Space (%v), remove and create new one\n", reflect.TypeOf(obj), obj.getPosition())
			}
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
	maxBomb int
	curBomb int
	score   int
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
	//drawLine(0, 0, fmt.Sprintf("counter=%v move=%v bombs=%v", count, state.speed, state.curBomb))
	state.screen.printString(&Dot{0, 0}, fmt.Sprintf("counter=%v move=%v bombs=%v", count, state.speed, state.curBomb))
	state.screen.printString(&Dot{0, state.screen.Height - 1}, fmt.Sprintf("score=%v", state.score))
	x, y := state.screen.Width/2+1, state.screen.Height/2+1
	for i := 0; i < state.maxBomb-state.curBomb; i++ {
		state.screen.printString(&Dot{x, y}, "**")
		state.screen.printString(&Dot{x, y + 1}, "**")
		x += 3
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
				pline := ""
				//fmt.Printf("str=%v\n", str)
				for idx, line := range strings.Split(str, "\n") {
					if idx == 0 {
						pline = line
					} else {
						d.buff[d.curLine] = pline //Todo: bad performance
						d.curLine = (d.curLine + 1) % len(d.buff)
						pline = line
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
			//upMsgs := state.space.move(time.Now(), forward, ctx, cancel)
			upMsgs := state.move(state.space, now, forward, ctx, cancel)
			//state.score += state.judgeExplosion(now, ctx, cancel)
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
