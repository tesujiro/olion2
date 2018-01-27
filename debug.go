package olion

import (
	"context"
	"fmt"
	"strings"
)

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
