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
	writeChan    chan string
	writeDone    chan struct{}
	readInitReq  chan struct{}
	readInitDone chan struct{}
	readNextReq  chan struct{} // reverse from current line
	readChan     chan string
	//readDone  chan struct{}
}

var debug *debugWriter

func newDebugWriter(ctx context.Context, cancel func()) {
	size := 1000
	d := &debugWriter{
		buff:         make([]string, size),
		curLine:      0,
		writeChan:    make(chan string),
		writeDone:    make(chan struct{}),
		readInitReq:  make(chan struct{}),
		readInitDone: make(chan struct{}),
		readNextReq:  make(chan struct{}),
		readChan:     make(chan string),
		//readDone:  make(chan struct{}),
	}

	go func() {
		defer cancel()
	L:
		for {
			var readNextLine int
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
			//case size := <-d.readReq:
			case <-d.readInitReq:
				readNextLine = d.curLine
				d.readInitDone <- struct{}{}
				/*
					firstLine := (d.curLine + len(d.buff) - size) % len(d.buff)
					for i := 0; i < size; i++ {
						idx := (firstLine + i) % len(d.buff)
						msg := d.buff[idx]
						//msg := strconv.Itoa(idx) + ":" + msg
						d.readChan <- msg
					}
				*/
			case <-d.readNextReq:
				msg := d.buff[readNextLine]
				readNextLine = (readNextLine - 1 + len(d.buff)) % len(d.buff)
				d.readChan <- fmt.Sprintf("%v:", readNextLine) + msg
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
	/*
		d.readReq <- w.height
		for i := 0; i < w.height; i++ {
			msg := <-d.readChan
			//msg = strconv.Itoa(i) + ":" + msg
			if len(msg) > w.width {
				msg = msg[:w.width]
			}
			for _, line := range sort.Reverse(CutStringInWidth(msg, w.width)) {
				w.screen.printString(&Dot{w.StartX, w.StartY + i}, line)
			}
		}
	*/
	reverse := func(strs []string) []string {
		for i := 0; i < len(strs)/2; i++ {
			j := len(strs) - i - 1
			strs[i], strs[j] = strs[j], strs[i]
		}
		return strs
	}

	d.readInitReq <- struct{}{}
	<-d.readInitDone
label:
	for i := w.height - 1; ; {
		if i < 0 {
			break label
		}
		d.readNextReq <- struct{}{}
		msg := <-d.readChan
		for _, line := range reverse(CutStringInWidth(msg, w.width)) {
			w.screen.printString(&Dot{w.StartX, w.StartY + i}, line)
			i--
			if i < 0 {
				break label
			}
		}
	}
}
