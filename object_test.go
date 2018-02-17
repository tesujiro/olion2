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
