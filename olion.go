package olion

import (
	"fmt"
	"syscall"
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

	fmt.Printf("Xpixel=%v Ypixel=%v\n", ws.Xpixel, ws.Ypixel)
	return uint(ws.Col), uint(ws.Row)
}

func initScreen() *Screen {
	w, h := getWinsize()
	return &Screen{Width: int(w), Height: int(h)}
}

func (sc *Screen) draw(d1, d2 Dot) {
}

func Run() {
	sc := initScreen()

	fmt.Printf("width=%v height=%v\n", sc.Width, sc.Height)
}
