package olion

import (
	"context"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

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

//func (state *Olion) move(spc *Space, t time.Time, dp Coordinates, ctx context.Context, cancel func()) []upMessage {
func (spc *Space) vanish(obj Exister, ctx context.Context, cancel func()) {
	//if fmt.Sprintf("%v", reflect.TypeOf(obj)) != "*olion.Star" {
	//debug.Printf("objct(%v) is out of the Space (%v), remove and create new one\n", reflect.TypeOf(obj), obj.getPosition())
	//}
	spc.deleteObj(obj)
	if !obj.isBomb() {
		//debug.Printf("objct is not a bomb\n")
		newObj := spc.GenFunc(time.Now())
		spc.addObj(newObj)
		go newObj.run(ctx, cancel)
		/*
			newObj.downCh() <- downMsg
			upMsg := <-newObj.upCh()
			upMsgs = append(upMsgs, upMsg)
		*/
	}
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
	//case true:
	//return newSpaceShip(now, 500, spc.randomSpace())
	//return newFramedRectangle(now, 1000, spc.randomSpace())
	case num < 2:
		return newBigShip(now, spc.randomSpace())
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

func NewSpace(ctx context.Context, cancel func(), objects int) *Space {
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
	for i := 0; i < objects; i++ {
		//for i := 0; i < 3; i++ {
		obj := spc.GenFunc(now)
		spc.addObj(obj)
		go obj.run(ctx, cancel)
	}

	return spc
}

func NewOuterSpace(ctx context.Context, cancel func(), objects int) *Space {
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
	for i := 0; i < objects; i++ {
		obj := spc.GenFunc(now)
		spc.addObj(obj)
		go obj.run(ctx, cancel)
	}

	//fmt.Printf("OuterSpace ==> %v Objects\n", len(spc.Objects))
	return spc
}
