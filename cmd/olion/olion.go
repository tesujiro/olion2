package main

import (
	"fmt"

	"github.com/tesujiro/olion"
)

func main() {
	sc := olion.initScreen()

	fmt.Println("width=%v height=%v\n", sc.Width, sc.Height)
}
