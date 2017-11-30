package olion

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Color struct {
	ColorId   int    `json:"colorId"`
	HexString string `json:"hexString"`
	RGB       RGB    `json:"rgb"`
	HSL       HSL    `json:"hsl"`
	Name      string `json:"name"`
}

type RGB struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

type HSL struct {
	H int `json:"h"`
	S int `json:"s"`
	L int `json:"l"`
}

func InitColor() {
	raw, err := ioutil.ReadFile("./color.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var colors []Color

	json.Unmarshal(raw, &colors)

	/*
		for _, c := range colors {
			fmt.Printf("color=%v\n", c)
		}
	*/
}
