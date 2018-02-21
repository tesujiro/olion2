package olion

import (
	"testing"
	"time"
)

func TestStar(t *testing.T) {
	cases := []struct {
		size          int
		c             Coordinates
		expectedSpeed Coordinates
		expectedParts int
	}{
		{size: 10, c: Coordinates{10, 20, 30}, expectedSpeed: Coordinates{0, 0, 0}, expectedParts: 1},
	}
	for _, c := range cases {
		o := newStar(time.Now(), c.size, c.c)
		actualSpeed := o.getSpeed()
		if actualSpeed != c.expectedSpeed {
			t.Errorf("got %v\nwant %v", actualSpeed, c.expectedSpeed)
		}
		actualParts := len(o.getParts())
		if actualParts != c.expectedParts {
			t.Errorf("got %v\nwant %v", actualParts, c.expectedParts)
		}
	}
}

func TestBomb(t *testing.T) {
	cases := []struct {
		size            int
		position, speed Coordinates
		expectedSpeed   Coordinates
		expectedParts   int
	}{
		{size: 10, position: Coordinates{10, 20, 30}, speed: Coordinates{1, 2, 3}, expectedSpeed: Coordinates{-1, -2, -3}, expectedParts: 1},
	}
	for _, c := range cases {
		o := newBomb(time.Now(), c.size, c.position, c.speed)
		actualSpeed := o.getSpeed()
		if actualSpeed != c.expectedSpeed {
			t.Errorf("got %v\nwant %v", actualSpeed, c.expectedSpeed)
		}
		actualParts := len(o.getParts())
		if actualParts != c.expectedParts {
			t.Errorf("got %v\nwant %v", actualParts, c.expectedParts)
		}
	}
}

func TestSpaceBox(t *testing.T) {
	cases := []struct {
		size                               int
		position                           Coordinates
		expectedSpeedMin, expectedSpeedMax int
		expectedParts                      int
	}{
		{size: 10, position: Coordinates{10, 20, 30}, expectedSpeedMin: 0, expectedSpeedMax: 40, expectedParts: 3},
	}
	for _, c := range cases {
		o := newBox(time.Now(), c.size, c.position)
		actualSpeed := o.getSpeed()
		if actualSpeed.Z < c.expectedSpeedMin {
			t.Errorf("got %v\nwant min:%v", actualSpeed, c.expectedSpeedMin)
		}
		if c.expectedSpeedMax < actualSpeed.Z {
			t.Errorf("got %v\nwant max:%v", actualSpeed, c.expectedSpeedMax)
		}
		actualParts := len(o.getParts())
		if actualParts != c.expectedParts {
			t.Errorf("got %v\nwant %v", actualParts, c.expectedParts)
		}
	}
}

func TestSpaceShip(t *testing.T) {
	cases := []struct {
		size                               int
		position                           Coordinates
		expectedSpeedMin, expectedSpeedMax int
		expectedParts                      int
	}{
		{size: 10, position: Coordinates{10, 20, 30}, expectedSpeedMin: 0, expectedSpeedMax: 40, expectedParts: 10},
	}
	for _, c := range cases {
		o := newSpaceShip(time.Now(), c.size, c.position)
		actualSpeed := o.getSpeed()
		if actualSpeed.Z < c.expectedSpeedMin {
			t.Errorf("got %v\nwant min:%v", actualSpeed, c.expectedSpeedMin)
		}
		if c.expectedSpeedMax < actualSpeed.Z {
			t.Errorf("got %v\nwant max:%v", actualSpeed, c.expectedSpeedMax)
		}
		actualParts := len(o.getParts())
		if actualParts != c.expectedParts {
			t.Errorf("got %v\nwant %v", actualParts, c.expectedParts)
		}
	}
}
