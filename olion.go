package olion

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
	"github.com/pkg/errors"
)

type ignorable interface {
	Ignorable() bool
}

type errIgnorable struct {
	err error
}

func (e errIgnorable) Ignorable() bool { return true }
func (e errIgnorable) Cause() error {
	return e.err
}
func (e errIgnorable) Error() string {
	return e.err.Error()
}
func makeIgnorable(err error) error {
	return &errIgnorable{err: err}
}

type Olion struct {
	Argv        []string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	Debug       bool
	Palette     bool
	Objects     int
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
	return &Olion{
		Argv:    os.Args,
		Stderr:  os.Stderr,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Pause:   false,
		readyCh: make(chan struct{}),
	}
}

func (state *Olion) Setup(ctx context.Context, cancel func()) error {
	rand.Seed(time.Now().UnixNano())

	//setup Options
	err := state.setupOptions()
	if err != nil {
		return errors.Wrap(err, "Setup Options Failed")
	}

	//setup Screen
	state.screen = NewScreen()

	//setup DebugWriter and DebugWindow
	buffSize := 1000
	newDebugWriter(buffSize, ctx, cancel)
	state.debugWindow = newDebugWindow(state.screen)

	//setup Space
	state.space = NewSpace(ctx, cancel, state.Objects)
	state.outerSpace = NewOuterSpace(ctx, cancel, 10)

	//setup Self Object
	state.position = Coordinates{X: 0, Y: 0, Z: 0}
	state.mobile = mobile{speed: Coordinates{X: 0, Y: 0, Z: 20}, time: time.Now()}
	state.maxBomb = 4
	state.curBomb = 0
	state.score = 0

	return nil
}

const version = "v0.0.1"

func (state *Olion) setupOptions() error {
	debug := flag.Bool("d", false, "Debug Mode")
	palette := flag.Bool("p", false, "Color Palette Mode")
	objects := flag.Int("o", 10, "Number of Flying Objects")
	help := flag.Bool("h", false, "Show Helps")
	ver := flag.Bool("v", false, "Show Version")
	flag.Parse()

	if *help {
		state.Stdout.Write([]byte("Help Messages"))
		//return makeIgnorable(errors.New("user asked to show help message"))
		return makeIgnorable(errors.New("user asked to show help message"))
	}

	if *ver {
		state.Stdout.Write([]byte("Olion version " + version + "\n"))
		return makeIgnorable(errors.New("user asked to show version"))
	}

	state.Debug = *debug
	state.Palette = *palette
	state.Objects = *objects
	return nil
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

	start := Dot{0, state.screen.Height - 5}
	disp_string(start, fmt.Sprintf("SCORE:%v", state.score))
	if state.Debug {
		state.screen.printString(&Dot{0, 0}, fmt.Sprintf("%v frameRate=%vfps counter=%v move=%v bombs=%v", time.Unix(state.dispFpsUnix, 0), state.dispFps, count, state.speed, state.curBomb))
		start = Dot{state.screen.Width - 30, state.screen.Height - 5}
		disp_string(start, fmt.Sprintf("%vFPS", state.dispFps))
	}

	x, y := state.screen.Width/2+1, state.screen.Height/2+1
	for i := 0; i < state.maxBomb-state.curBomb; i++ {
		state.screen.printString(&Dot{x, y}, "**")
		state.screen.printString(&Dot{x, y + 1}, "**")
		x += 3
		y += 0
	}
}

func (state *Olion) setStatus() {
	//count bombs
	bombs := 0
	//for _, obj := range state.space.Objects {
	for _, obj := range state.space.GetObjects() {
		if obj.isBomb() {
			bombs++
		}
	}
	state.curBomb = bombs
}

func (state *Olion) move(spc *Space, t time.Time, dp Coordinates) []upMessage {
	downMsg := downMessage{
		time:          t,
		deltaPosition: dp,
	}
	upMsgs := []upMessage{}
	now := time.Now()

	// send downMsg , and get flying objects and bombs
	flyings := []Exister{}
	bombs := []Exister{}
	for _, obj := range spc.GetObjects() {
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
			state.space.AddObj(newObj)
			flying.removeBomb()
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

	for _, obj := range spc.GetObjects() {
		// stop flying object explosion
		if obj.isExploding() {
			// Delete 10 sec. after explosion.
			deltaTime := float64(time.Now().Sub(obj.getExplodedTime()) / time.Millisecond)
			if deltaTime > float64(1e4) {
				spc.Vanish(obj)
			}
		}
		// if objct is out of the Space , remove it and create new one
		if !spc.inTheSpace(obj.getPosition()) {
			spc.Vanish(obj)
		}
	}

	return upMsgs
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
			upMsgsOuterSpace := state.move(state.outerSpace, time.Now(), speed)
			view.draw(upMsgsOuterSpace)
			//Space
			now := time.Now()
			if fireBomb {
				debug.Printf("newBomb\n")
				speed := Coordinates{state.speed.X, state.speed.Y, state.speed.Z + 80}
				newObj := newBomb(now, 1000, Coordinates{}, speed)
				state.space.AddObj(newObj)
				fireBomb = false
			}
			forward := state.getDistance(now)
			upMsgs := state.move(state.space, now, forward)
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

	err = state.Setup(ctx, cancel)
	if err != nil {
		return errors.Wrap(err, "Setup Failed")
	}

	go state.Loop(NewView(state), ctx, cancel)

	// Alright, done everything we need to do automatically. We'll let
	// the user play with peco, and when we receive notification to
	// bail out, the context should be canceled appropriately
	<-ctx.Done()

	return nil
}
