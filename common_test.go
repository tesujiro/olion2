package olion

import "testing"

func TestAdd(t *testing.T) {
	cases := []struct {
		c1, c2, expected Coordinates
	}{
		{c1: Coordinates{0, 0, 0}, c2: Coordinates{0, 0, 0}, expected: Coordinates{0, 0, 0}},
		{c1: Coordinates{1, 2, 3}, c2: Coordinates{10, 20, 30}, expected: Coordinates{11, 22, 33}},
	}
	for _, c := range cases {
		actual := c.c1.Add(c.c2)
		if actual != c.expected {
			t.Errorf("got %v\nwant %v", actual, c.expected)
		}
	}
}
