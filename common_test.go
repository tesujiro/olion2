package olion

import (
	"reflect"
	"testing"
)

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

func TestScaleBy(t *testing.T) {
	cases := []struct {
		c, expected Coordinates
		k           int
	}{
		{c: Coordinates{0, 0, 0}, k: 1, expected: Coordinates{0, 0, 0}},
		{c: Coordinates{1, 2, 3}, k: 0, expected: Coordinates{0, 0, 0}},
		{c: Coordinates{1, 2, 3}, k: 2, expected: Coordinates{2, 4, 6}},
		{c: Coordinates{1, 2, 3}, k: -3, expected: Coordinates{-3, -6, -9}},
	}
	for _, c := range cases {
		actual := c.c.ScaleBy(c.k)
		if actual != c.expected {
			t.Errorf("got %v\nwant %v", actual, c.expected)
		}
	}
}

func TestDiv(t *testing.T) {
	cases := []struct {
		c, expected Coordinates
		k           int
	}{
		{c: Coordinates{0, 0, 0}, k: 1, expected: Coordinates{0, 0, 0}},
		{c: Coordinates{1, 2, 3}, k: 1, expected: Coordinates{1, 2, 3}},
		{c: Coordinates{1, 2, 3}, k: -1, expected: Coordinates{-1, -2, -3}},
		{c: Coordinates{4, 8, 12}, k: 2, expected: Coordinates{2, 4, 6}},
	}
	for _, c := range cases {
		actual := c.c.Div(c.k)
		if actual != c.expected {
			t.Errorf("got %v\nwant %v", actual, c.expected)
		}
	}
}

func TestSymmetry(t *testing.T) {
	cases := []struct {
		center, diff Coordinates
		expected     []Coordinates
	}{
		{center: Coordinates{0, 0, 0}, diff: Coordinates{0, 2, 3},
			expected: []Coordinates{Coordinates{0, 2, 3}, Coordinates{0, 2, -3}, Coordinates{0, -2, -3}, Coordinates{0, -2, 3}}},
		{center: Coordinates{0, 0, 0}, diff: Coordinates{1, 0, 3},
			expected: []Coordinates{Coordinates{1, 0, 3}, Coordinates{1, 0, -3}, Coordinates{-1, 0, -3}, Coordinates{-1, 0, 3}}},
		{center: Coordinates{0, 0, 0}, diff: Coordinates{1, 2, 0},
			expected: []Coordinates{Coordinates{1, 2, 0}, Coordinates{1, -2, 0}, Coordinates{-1, -2, 0}, Coordinates{-1, 2, 0}}},
	}
	for _, c := range cases {
		actual := c.center.Symmetry(c.diff)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("got %v\nwant %v", actual, c.expected)
		}
	}
}
