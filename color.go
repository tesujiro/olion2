package olion

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	termbox "github.com/nsf/termbox-go"
)

type Color struct {
	ColorId   Attribute `json:"colorId"`
	HexString string    `json:"hexString"`
	RGB       RGB       `json:"rgb"`
	HSL       HSL       `json:"hsl"`
	Name      string    `json:"name"`
}

func (color Color) Attribute() Attribute {
	return color.ColorId + 1
}

type Colors []Color

func (colors Colors) name(n string) Color {
	for i, c := range colors {
		if c.Name == n {
			return colors[i]
		}
	}
	return colors.name("White")
}

type RGB struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

type HSL struct {
	H float64 `json:"h"`
	S int     `json:"s"`
	L int     `json:"l"`
}

var colors Colors

func InitColor() {
	raw, err := ioutil.ReadFile("./color.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = json.Unmarshal(raw, &colors)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func (state *Olion) drawColorPalette() {
	//debug.Printf("colors=%v\n", len(colors))
	//var longest Color
	//for _, col := range colors {
	//	debug.Printf("id=%v\tname=%v\trgb=%v\n", col.ColorId, col.Name, col.RGB)
	//	if len(col.Name) > len(longest.Name) {
	//		longest = col
	//	}
	//}
	//debug.Printf("LONGEST id=%v\tname=%v\trgb=%v\tlen(name)=%v\n", longest.ColorId, longest.Name, longest.RGB, len(longest.Name))
	colorsPerLine := 13
	length := 16
	interval := 0
	attributes := 3
	for idx, col := range colors {
		startX := (idx % colorsPerLine) * (length + interval)
		startY := (idx / colorsPerLine) * (attributes + interval)
		state.screen.printString(&Dot{startX, startY}, strconv.Itoa(int(col.ColorId)))
		state.screen.printStringWithColor(&Dot{startX, startY + 1}, col.Name, termbox.Attribute(col.ColorId)+1)
		for x := 0; x < length; x++ {
			termbox.SetCell(startX+x, startY+2, ' ', termbox.ColorWhite, termbox.Attribute(col.ColorId)+1)
		}
	}

}
