package olion

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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

	//for _, c := range colors {
	//fmt.Printf("color=%v\n", c)
	//}
	//fmt.Printf("%v\n", colors.name("Black"))
	//fmt.Printf("%v\n", colors.name("White"))
	//fmt.Printf("%v\n", colors.name("Grey"))
	//fmt.Printf("ColorId=%v\n", colors.name("Grey").ColorId)
}
