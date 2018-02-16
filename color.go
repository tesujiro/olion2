package olion

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	termbox "github.com/nsf/termbox-go"
	"github.com/pkg/errors"
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

func InitColor() error {
	raw, err := ioutil.ReadFile("./color.json")
	if err != nil {
		return errors.Wrap(err, "Read ./color.json")
	}

	err = json.Unmarshal(raw, &colors)
	if err != nil {
		return errors.Wrap(err, "Unmarshal ./color.json")
	}
	return nil
}

func (state *Olion) drawColorPalette() {
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
